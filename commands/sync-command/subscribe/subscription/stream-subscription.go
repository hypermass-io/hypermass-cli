package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/helpers"
	"hypermass-cli/commands/sync-command/subscribe"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	subscription_status "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-status"
	"hypermass-cli/config"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Subscription struct {
	StreamId  string
	ParentCtx context.Context
	Ctx       context.Context
	Cancel    context.CancelFunc

	Auth                      config.HypermassAuth
	SubscriptionConfiguration config.SubscriptionConfiguration

	FolderPath    string
	LastPayloadId string

	FileQueue chan *messages.PayloadNotificationMessage

	Writer     payload_writers.PayloadWriterStrategy
	StartPoint string

	// RequestRestart is called once, to requests that this subscription be stopped, and a replacement created after a delay
	RequestRestart func(streamId string, err error)
	stopOnce       sync.Once

	//ProcessorsWG a WG tracking the processors such as RetryingInfoChannelSubscription and StartFileQueueProcessor
	ProcessorsWG sync.WaitGroup

	ReportingState subscription_status.SubscriptionReportingState
}

func NewSubscription(
	parentCtx context.Context,
	streamConfig config.SubscriptionConfiguration,
	auth config.HypermassAuth,
	startupDelay time.Duration,
) (*Subscription, error) {

	ctx, cancel := context.WithCancel(parentCtx)

	folderPath := helpers.GetStreamPathFromConfig(streamConfig.TargetDirectory)
	directoryError := subscriptionhelpers.InitialiseAndCheckDirectory(folderPath)
	lastPayloadId := subscriptionhelpers.ReadLastPayloadId(folderPath)

	if directoryError != nil {
		log.Println(directoryError)
		log.Println("Unable to initialise directory")
		cancel()
		return nil, directoryError
	}

	subscription := &Subscription{
		StreamId:                  streamConfig.Key,
		ParentCtx:                 parentCtx,
		Ctx:                       ctx,
		Cancel:                    cancel,
		Auth:                      auth,
		SubscriptionConfiguration: streamConfig,
		FolderPath:                folderPath,
		LastPayloadId:             lastPayloadId,
		FileQueue:                 make(chan *messages.PayloadNotificationMessage, 100000),
		Writer:                    payload_writers.GetPayloadWriter(streamConfig.WriterType, streamConfig.Key),
		StartPoint:                streamConfig.StartPoint,
		ReportingState:            subscription_status.NewInitialState(startupDelay),
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Println("Subscription stopped before async pollers started: ", streamConfig.Key)
			subscription.ReportingState = subscription_status.NewStoppedState()
			return
		case <-time.After(startupDelay):
			subscription.ReportingState = subscription_status.NewConnectingState()
			//start async subscriber processes
			subscription.ProcessorsWG.Go(subscription.StartFileQueueProcessor)
			subscription.ProcessorsWG.Go(subscription.RetryingInfoChannelSubscription)
		}
	}()

	return subscription, nil
}

// indicates that the subscription is broken and should stop (later to be retried)
func (s *Subscription) restartSubscriptionWithReason(err error) {
	s.ReportingState = subscription_status.NewRestartingState()

	s.stopOnce.Do(func() {
		if s.RequestRestart != nil {
			s.RequestRestart(s.SubscriptionConfiguration.Key, err)
		}
	})
}

func (s *Subscription) RetryingInfoChannelSubscription() {
	//connect and receive messages
	err := s.startInfoChannelReader()

	//only returns if interrupted
	fmt.Println("Subscription exited for " + s.StreamId)

	//determine if the main process is cancelled or if the executor failed
	select {
	case <-s.Ctx.Done():
		//exit if the parent context is done, no other action needed
		return
	default:
		//otherwise trigger the retry if the connection is lost
		s.restartSubscriptionWithReason(err)
	}
}

// startInfoChannelReader a connection to the infochannel
// This can be interrupted, in which case it will return without effecting the context
func (s *Subscription) startInfoChannelReader() error {
	fmt.Println("Started subscription to stream: " + s.StreamId)

	//make the initial http request to get the signed websocket URL
	signedWebsocketUrl, authErr := subscribe.GetAuthorizedSubscriptionUrl(
		s.Auth, s.SubscriptionConfiguration.Key, s.LastPayloadId)

	if authErr != nil {
		return authErr
	}

	//connect to the websocket
	websocketUrl, websocketParseError := url.Parse(signedWebsocketUrl)
	if websocketParseError != nil {
		log.Println(websocketParseError)
		log.Println("Bad internal URL, please report this to support")
		return websocketParseError
	}

	websocketConnection, _, websocketError := websocket.DefaultDialer.Dial(websocketUrl.String(), nil)
	if websocketError != nil {
		log.Printf("Unable to connect to info channel, cause: %v", websocketError)
		return &app_errors.ConnectionLostError{Message: "Unable to connect to info channel"}
	}

	defer func() {
		if websocketConnection != nil {
			_ = websocketConnection.Close()
		}
	}()

	// This goroutine sits and waits specifically for the shutdown signal.
	// when/if the context is Done, it "kicks" the websocket so the reader loop (below) unblocks.
	stopWatcher := make(chan struct{})
	go func() {
		select {
		case <-s.Ctx.Done():
			log.Printf("Closing websocket for %s due to context cancellation", s.StreamId)
			_ = websocketConnection.Close() // This triggers ReadMessage to return err
		case <-stopWatcher:
			// The reader finished naturally, no need to force close
			return
		}
	}()
	defer close(stopWatcher)

	s.ReportingState = subscription_status.NewConnectedActivityState()
	//blocking loop of messages being processed from the websocket
	for {
		//ReadMessage is blocking
		_, message, err := websocketConnection.ReadMessage()
		if err != nil {
			log.Println("read message error:", err)
			return err
		}

		messageType := determineMessageType(message)

		switch messageType {
		case "PayloadNotificationMessage":
			s.ReportingState = subscription_status.NewConnectedActivityState()
			err2, done := s.writePayloadMessageToQueue(message)
			if done {
				return err2
			}
		case "PingPong":
			s.ReportingState = subscription_status.NewConnectedActivityState()
			_ = s.writePongResponse(websocketConnection)
		}

	}
}

func (s *Subscription) writePongResponse(websocketConnection *websocket.Conn) error {
	pong := messages.PingPongResponseMessage{
		Type: "Pong",
	}

	pongBytes, err := json.Marshal(pong)
	if err != nil {
		log.Printf("failed to marshal pong response for %s: %v", s.StreamId, err)
		return err
	}

	if writeErr := websocketConnection.WriteMessage(websocket.TextMessage, pongBytes); writeErr != nil {
		log.Printf("failed to write pong response for %s: %v", s.StreamId, writeErr)
		return writeErr
	}
	return nil
}

func (s *Subscription) writePayloadMessageToQueue(message []byte) (error, bool) {
	data := messages.PayloadNotificationMessage{}

	messageErr := json.Unmarshal(message, &data)
	if messageErr != nil {
		log.Println("unmarshalling notification error:", messageErr)
		return messageErr, true
	}

	// Write to the queue, checking for cancellation while blocked
	select {
	case <-s.Ctx.Done():
		// Cancelled while trying to write to the filequeue
		return nil, true
	case s.FileQueue <- &data:
		// Success, no action needed (loop)
	}
	return nil, false
}

func (s *Subscription) StartFileQueueProcessor() {
	for {
		select {
		case msg := <-s.FileQueue:
			// only respond to known message types
			if msg.Type == "PayloadNotificationMessage" {
				// Process the message
				fmt.Printf("Downloading payload %s for stream %s \n", msg.PayloadId, msg.StreamId)

				downloadPayloadErr := subscriptionhelpers.DownloadPayload(s.Auth, s.FolderPath, s.Writer, *msg)
				if downloadPayloadErr != nil {
					log.Printf("Failed to download: %s\n", downloadPayloadErr)
					s.restartSubscriptionWithReason(downloadPayloadErr)
					return //exit the "StartFileQueueProcessor" loop completely - this instance won't recover
				}

				writeEtagErr := subscriptionhelpers.WriteLastPayloadId(s.FolderPath, msg.PayloadId)
				if writeEtagErr != nil {
					log.Println("Failed to record the last payload id (may result in repeated message): ", writeEtagErr)
					s.restartSubscriptionWithReason(writeEtagErr)
					return //exit the "StartFileQueueProcessor" loop completely - this instance won't recover
				}

				fmt.Printf("Received payload %s for stream %s \n", msg.PayloadId, msg.StreamId)
				s.LastPayloadId = msg.PayloadId
			}

		case <-s.Ctx.Done():
			fmt.Println("Connection to stream stopped: " + s.StreamId)
			return
		}
	}
}

func determineMessageType(message []byte) string {
	data := messages.GenericTypedMessage{}

	messageErr := json.Unmarshal(message, &data)
	if messageErr != nil {
		log.Println("unmarshalling notification error:", messageErr)
		return "unknown"
	}

	return data.Type
}
