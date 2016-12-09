package voice

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

type Endpoints struct {
	VoiceEndpoint endpoint.Endpoint
}

// Voice Endpoint
func (e Endpoints) Voice(ctx context.Context, voice VoiceRequest) (*VoiceResponse, error) {
	response, err := e.VoiceEndpoint(ctx, voice)
	return response.(*VoiceResponse), err
}

func MakeVoiceEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (response interface{}, err error) {
		request := req.(VoiceRequest)
		return s.Voice(ctx, request)
	}
}
