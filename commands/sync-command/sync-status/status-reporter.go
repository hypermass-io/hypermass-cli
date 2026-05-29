package sync_status

import (
	"encoding/json"
	"hypermass-cli/app_common"
	"hypermass-cli/commands/sync-command/publish/publication"

	"hypermass-cli/commands/sync-command/subscribe/subscription"
)

func ReportStatus(subscriptionPollersMap map[string]*subscription.Subscription, publicationPollersMap map[string]*publication.PublicationPoller) (string, error) {
	subscriptionStatuses := buildSubscriptionStatuses(subscriptionPollersMap)
	publicationStatuses := buildPublicationStatuses(publicationPollersMap)

	report := app_common.StatusReport{
		Subscriptions: subscriptionStatuses,
		Publications:  publicationStatuses,
	}

	jsonBytes, err := json.Marshal(report)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func buildPublicationStatuses(p map[string]*publication.PublicationPoller) []app_common.PublicationStatus {
	statuses := make([]app_common.PublicationStatus, 0, len(p))
	for _, reportPublication := range p {
		statuses = append(statuses, app_common.PublicationStatus{
			StreamId:         reportPublication.StreamId,
			Status:           reportPublication.ReportingState.Status,
			Description:      reportPublication.ReportingState.Description,
			PollWaitDuration: reportPublication.ReportingState.PollWaitDuration,
			NextPoll:         reportPublication.ReportingState.NextPoll,
			QueueSize:        reportPublication.ReportingState.QueueSize,
		})
	}

	return statuses
}

func buildSubscriptionStatuses(s map[string]*subscription.Subscription) []app_common.SubscriptionStatus {
	statuses := make([]app_common.SubscriptionStatus, 0, len(s))
	for _, reportSubscription := range s {
		statuses = append(statuses, app_common.SubscriptionStatus{
			StreamId:       reportSubscription.StreamId,
			LastPayloadId:  reportSubscription.LastPayloadId,
			FetchQueueSize: len(reportSubscription.FileQueue),
			Status:         reportSubscription.ReportingState.Status,
			Description:    reportSubscription.ReportingState.Description,
			WaitDuration:   reportSubscription.ReportingState.WaitDuration,
			WaitingUntil:   reportSubscription.ReportingState.WaitingUntil,
			LastActivity:   reportSubscription.ReportingState.LastActivity,
		})
	}

	return statuses
}
