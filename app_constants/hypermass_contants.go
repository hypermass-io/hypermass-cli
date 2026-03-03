package app_constants

// BulkAuthenticationApiUrl is the first URL called in interactions with the Hypermass API - it will
var BulkAuthenticationApiUrl string = "https://auth.hypermass.io/api/data/bulk/authorise/infochannel"
var PublicApiUrl string = "https://api.hypermass.io/api"

// These will be populated by the linker at build time
var HypermassCliVersion string = "v0.0.0-dev"
var Commit string = "none"
var BuildDate string = "unknown"

// GetBulkAuthenticationApiUrl constructs the full URL using the base URL and streamId.
func GetBulkAuthenticationApiUrl(streamId string, lastPayload string) string {
	if lastPayload != "" {
		return BulkAuthenticationApiUrl + "/" + streamId + "?lastPayload=" + lastPayload
	} else {
		return BulkAuthenticationApiUrl + "/" + streamId
	}
}
