package generated

type CodelabResponse struct {
	Files Files `json:"files,omitempty"`
}

type Files struct {
	Payload Payload `json:"payload,omitempty"`
}

type Payload struct {
	Signs map[string]map[string]any `json:"signs,omitempty"`
}
