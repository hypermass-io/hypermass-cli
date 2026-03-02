//go:build local

package app_constants

func init() {
	BulkAuthenticationApiUrl = "http://localhost:9000/api/data/bulk/authorise/infochannel"
	PublicApiUrl = "http://localhost:9000/api"
}
