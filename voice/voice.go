package voice

type VoiceRequest struct {
	Audio       []byte
	SampleCount uint32
}

type VoiceResponse struct {
	Code int         `json:"code"`
	Body interface{} `json:"body"`
}
