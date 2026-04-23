package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"hypermass-cli/commands/sync-command/helpers"
	"hypermass-cli/commands/sync-command/subscribe"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
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

	subscription := Subscription{
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
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Println("Subscription stopped before async pollers started: ", streamConfig.Key)
			return
		case <-time.After(startupDelay):
			//start async subscriber processes
			subscription.ProcessorsWG.Go(subscription.StartFileQueueProcessor)
			subscription.ProcessorsWG.Go(subscription.RetryingInfoChannelSubscription)
		}
	}()

	return &subscription, nil
}

// indicates that the subscription is broken and should stop (later to be retried)
func (s *Subscription) restartSubscriptionWithReason(err error) {
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
		log.Println(websocketError)
		log.Println("Unable to connect to info channel")
		return websocketError
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

	//blocking loop of messages being processed from the websocket
	for {
		//ReadMessage is blocking
		_, message, err := websocketConnection.ReadMessage()
		if err != nil {
			log.Println("read message error:", err)
			return err
		}

		data := messages.PayloadNotificationMessage{}

		messageErr := json.Unmarshal(message, &data)
		if messageErr != nil {
			log.Println("unmarshalling notification error:", err)
			return messageErr
		}

		// Write to the queue, checking for cancellation while blocked
		select {
		case <-s.Ctx.Done():
			// Cancelled while trying to write to the filequeue
			return nil
		case s.FileQueue <- &data:
			// Success, no action needed (loop)
		}
	}
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
			}

		case <-s.Ctx.Done():
			fmt.Println("Connection to stream stopped: " + s.StreamId)
			return
		}
	}
}
