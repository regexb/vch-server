package voice

type VoiceRequest struct {
	Audio []byte
}

type VoiceResponse struct {
	Code int         `json:"code"`
	Body interface{} `json:"body"`
}
