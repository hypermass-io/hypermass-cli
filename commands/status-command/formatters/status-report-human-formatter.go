package formatters

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/config/synclock"
	"sort"
	"time"
)

func FormatHumanReadableMessage(response *synclock.CommandResponse) {
	statusReport, err := deserializeStatusReport(response.Data)
	if err != nil {
		fmt.Printf("❌ Failed to read status response: %v\n", err)
		return
	}

	fmt.Println("✅ Sync is running")

	fmt.Println("\nSubscriptions:")
	if len(statusReport.Subscriptions) == 0 {
		fmt.Println("- none")
	} else {
		printSubscriptionStatusTable(statusReport.Subscriptions)
	}

	fmt.Println("\nPublications:")
	if len(statusReport.Publications) == 0 {
		fmt.Println("- none")
	} else {
		printPublicationStatusTable(statusReport.Publications)
	}
}

func printPublicationStatusTable(publications []app_common.PublicationStatus) {
	sort.Slice(publications, func(i, j int) bool {
		return publications[i].StreamId < publications[j].StreamId
	})

	fmt.Printf(
		"%-14s %-22s %10s %-21s %-12s\n",
		"Stream ID",
		"Status",
		"Queue Size",
		"Next Poll",
		"Time Remaining",
	)

	for _, publication := range publications {
		fmt.Printf(
			"%-14s %-22s %10d %-21s %-12s\n",
			truncate(publication.StreamId, 14),
			truncate(publication.Status, 22),
			publication.QueueSize,
			formatTimeOrDash(publication.NextPoll),
			formatDurationOrDash(time.Until(publication.NextPoll)),
		)
	}
}

func printSubscriptionStatusTable(subscriptions []app_common.SubscriptionStatus) {
	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].StreamId < subscriptions[j].StreamId
	})

	fmt.Printf(
		"%-14s %-22s %16s %-40s %-20s %-12s\n",
		"Stream ID",
		"Status",
		"Fetch Queue Size",
		"Last Payload Id",
		"Last Activity",
		"Time Remaining",
	)

	for _, subscription := range subscriptions {
		fmt.Printf(
			"%-14s %-22s %16d %-40s %-20s %-12s\n",
			truncate(subscription.StreamId, 14),
			truncate(subscription.Status, 22),
			subscription.FetchQueueSize,
			truncate(emptyOrDash(subscription.LastPayloadId), 40),
			formatTimeOrDash(subscription.LastActivity),
			formatDurationOrDash(time.Until(subscription.WaitingUntil)),
		)
	}
}

func formatTimeOrDash(value time.Time) string {
	if value.IsZero() {
		return "-"
	}

	return value.Format("2006-01-02 15:04:05")
}

func formatDurationOrDash(value time.Duration) string {
	if value <= 0 {
		return "-"
	}

	return value.Round(time.Second).String()
}

func emptyOrDash(value string) string {
	if value == "" {
		return "-"
	}

	return value
}

func truncate(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}

	if maxLength <= 3 {
		return value[:maxLength]
	}

	return value[:maxLength-3] + "..."
}

func deserializeStatusReport(data interface{}) (app_common.StatusReport, error) {
	var statusReport app_common.StatusReport

	jsonBytes, ok := data.(string)
	if !ok {
		return statusReport, fmt.Errorf("status response data was not a byte array")
	}

	if err := json.Unmarshal([]byte(jsonBytes), &statusReport); err != nil {
		return statusReport, err
	}

	return statusReport, nil
}
