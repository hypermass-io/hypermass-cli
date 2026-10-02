package publish_command

import (
	"errors"
	"fmt"
	"hypermass-cli/app_errors"
	reload_command "hypermass-cli/commands/reload-command"
	publication_helpers "hypermass-cli/commands/sync-command/publish/publication/publication-helpers"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"os"
	"path/filepath"
)

func Publish(streamId string) {
	configuration, err := config.ReadConfiguration()
	if err != nil {
		stop(fmt.Sprintf("❌ %v", err))
	}

	auth, err := config.ReadSecretKey()
	if err != nil {
		stop(fmt.Sprintf("❌ %v", err))
	}
	if !auth.HasKey() {
		stop("⚠️ Publishing needs a free account: add your access key with 'hypermass init'.")
	}

	for _, existing := range configuration.PublicationConfigurations {
		if existing.Key == streamId {
			fmt.Printf("✅ Already publishing to %s from %s\n", streamId, existing.TargetDirectory)
			return
		}
	}

	baseDirectory, err := configuration.BaseDirectoryOrDefault()
	if err != nil {
		stop(fmt.Sprintf("❌ %v", err))
	}

	if err := checkCanPublish(auth, streamId); err != nil {
		stop("❌ " + err.Error())
	}

	entry := config.PublicationConfiguration{
		Key:             streamId,
		TargetDirectory: filepath.Join(baseDirectory, "publications", streamId),
		DisposerType:    "delete-on-success",
	}

	if err := config.AddPublication(entry); err != nil {
		stop(fmt.Sprintf("❌ Could not add the publication: %v\nAdd it to publication-sources in hypermass-config.yaml by hand:\n"+
			"  - key: %s\n    target-directory: %s\n    disposer-type: %s", err, entry.Key, entry.TargetDirectory, entry.DisposerType))
	}

	fmt.Printf("✅ Publishing to %s, files placed in %s are published, then deleted.\n", streamId, entry.TargetDirectory)
	fmt.Println("To keep published files, set disposer-type to move-on-success in hypermass-config.yaml.")

	if _, _, err := synclock.DialSync(); err != nil {
		fmt.Println("Run 'hypermass sync' to start publishing.")
		return
	}

	reload_command.Reload(false)
}

// checkCanPublish asks the service whether the access key can publish to the stream, returning why not.
func checkCanPublish(auth config.HypermassAuth, streamId string) error {
	_, err := publication_helpers.GetConfigurationForStream(config.HypermassProfile{Auth: auth}, streamId)
	if err == nil {
		return nil
	}

	var notFound *app_errors.StreamNotFoundError
	var accessDenied *app_errors.StreamAccessDeniedError
	var credentialsRejected *app_errors.CredentialsRejectedError
	var connectionLost *app_errors.ConnectionLostError

	switch {
	case errors.As(err, &notFound):
		return fmt.Errorf("there is no stream with the id %s", streamId)
	case errors.As(err, &accessDenied):
		return fmt.Errorf("stream %s belongs to another account, you can publish to your own streams", streamId)
	case errors.As(err, &credentialsRejected):
		return errors.New("your access key was rejected, check it at https://hypermass.io/access-keys")
	case errors.As(err, &connectionLost):
		return fmt.Errorf("could not check stream %s (%s), please try again", streamId, connectionLost.Error())
	default:
		return fmt.Errorf("could not check stream %s (%v), please try again", streamId, err)
	}
}

func stop(message string) {
	fmt.Println(message)
	os.Exit(1)
}
