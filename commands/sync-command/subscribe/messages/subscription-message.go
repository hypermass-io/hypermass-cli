package messages

type InfoChannelMessage struct {
	ConnectionURL string `json:"connectionUrl"`
}

type GenericTypedMessage struct {
	Type string `json:"type"`
}

type PayloadNotificationMessage struct {
	Type                 string  `json:"type"`
	StreamId             string  `json:"streamId"`
	PayloadId            string  `json:"payloadId"`
	FileExtension        string  `json:"fileExtension"`
	PublishedTimestamp   string  `json:"publishedTimestamp"`
	BytesCount           int64   `json:"bytesCount"`
	ContentHashAlgorithm *string `json:"contentHashAlgorithm"` // nil for payloads stored before hashes were recorded
	ContentHash          *string `json:"contentHash"`          // base64
	ValidationType       string  `json:"validationType"`       // "basic" or "schema"
	DownloadUrl          string  `json:"downloadUrl"`
}

type PingPongMessage struct {
	Type string `json:"type"`
}
type PingPongResponseMessage struct {
	Type string `json:"type"`
}
