package subscribe_command

import (
	"errors"
	"fmt"
	"hypermass-cli/app_errors"
	reload_command "hypermass-cli/commands/reload-command"
	"hypermass-cli/commands/sync-command/subscribe"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"os"
	"path/filepath"
)

func Subscribe(streamId string) {
	configuration, err := config.ReadConfiguration()
	if err != nil {
		stop(fmt.Sprintf("❌ %v", err))
	}

	//without a key the subscription is anonymous, within the free daily allowance
	auth, err := config.ReadSecretKey()
	if err != nil {
		stop(fmt.Sprintf("❌ %v", err))
	}

	for _, existing := range configuration.SubscriptionConfigurations {
		if existing.Key == streamId {
			fmt.Printf("✅ Already subscribed to %s, files arrive in %s\n", streamId, existing.TargetDirectory)
			return
		}
	}

	baseDirectory, err := configuration.BaseDirectoryOrDefault()
	if err != nil {
		stop(fmt.Sprintf("❌ %v", err))
	}

	allowanceWarning, err := checkCanSubscribe(auth, streamId)
	if err != nil {
		stop("❌ " + err.Error())
	}

	entry := config.SubscriptionConfiguration{
		Key:             streamId,
		TargetDirectory: filepath.Join(baseDirectory, "subscriptions", streamId),
		WriterType:      "file-per-payload",
	}

	if err := config.AddSubscription(entry); err != nil {
		stop(fmt.Sprintf("❌ Could not add the subscription: %v\nAdd it to subscription-targets in hypermass-config.yaml by hand:\n"+
			"  - key: %s\n    target-directory: %s\n    writer-type: %s", err, entry.Key, entry.TargetDirectory, entry.WriterType))
	}

	fmt.Printf("✅ Subscribed to %s, files will arrive in %s\n", streamId, entry.TargetDirectory)
	if allowanceWarning != "" {
		fmt.Println("⚠️ " + allowanceWarning)
	}

	if _, _, err := synclock.DialSync(); err != nil {
		fmt.Println("Run 'hypermass sync' to start receiving.")
		return
	}

	reload_command.Reload(false)
}

// checkCanSubscribe asks the service whether the access key can subscribe to the stream, returning why not. An
// exhausted allowance still subscribes, returning a warning, as data flows again once the allowance resets.
func checkCanSubscribe(auth config.HypermassAuth, streamId string) (string, error) {
	_, err := subscribe.GetAuthorizedSubscriptionUrl(auth, streamId, "")
	if err == nil {
		return "", nil
	}

	var insufficientAllowance *app_errors.InsufficientAllowanceError
	var notFound *app_errors.StreamNotFoundError
	var accessDenied *app_errors.StreamAccessDeniedError
	var credentialsRejected *app_errors.CredentialsRejectedError
	var connectionLost *app_errors.ConnectionLostError

	switch {
	case errors.As(err, &insufficientAllowance):
		return "Your account has no allowance left, files will arrive once it resets.", nil
	case errors.As(err, &notFound):
		return "", fmt.Errorf("there is no stream with the id %s", streamId)
	case errors.As(err, &accessDenied):
		return "", fmt.Errorf("stream %s is not available to your access key", streamId)
	case errors.As(err, &credentialsRejected):
		return "", errors.New("your access key was rejected, check it at https://hypermass.io/access-keys")
	case errors.As(err, &connectionLost):
		return "", fmt.Errorf("could not check stream %s (%s), please try again", streamId, connectionLost.Error())
	default:
		return "", fmt.Errorf("could not check stream %s (%v), please try again", streamId, err)
	}
}

func stop(message string) {
	fmt.Println(message)
	os.Exit(1)
}
