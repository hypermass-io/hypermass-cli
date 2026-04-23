package messages

type InfoChannelMessage struct {
	ConnectionURL string `json:"connectionUrl"`
}

type GenericTypedMessage struct {
	Type string `json:"type"`
}

type PayloadNotificationMessage struct {
	Type               string `json:"type"`
	StreamId           string `json:"streamId"`
	PayloadId          string `json:"payloadId"`
	FileExtension      string `json:"fileExtension"`
	PublishedTimestamp string `json:"publishedTimestamp"`
	DownloadUrl        string `json:"downloadUrl"`
}

type PingPongMessage struct {
	Type string `json:"type"`
}
type PingPongResponseMessage struct {
	Type string `json:"type"`
}
