package app_common

// ReloadReport is the reply to a reload, listing each stream the reload touched.
type ReloadReport struct {
	Streams   []ReloadedStream `json:"streams"`
	Unchanged int              `json:"unchanged"`
	Warnings  []string         `json:"warnings"`

	// Completed is false when the reply was sent before the changes were applied
	Completed bool `json:"completed"`
}

type ReloadedStream struct {
	StreamId  string `json:"streamId"`
	Direction string `json:"direction"` // subscription or publication
	Change    string `json:"change"`    // added, changed or removed
	Error     string `json:"error,omitempty"`
}
