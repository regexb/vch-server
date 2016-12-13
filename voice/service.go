package voice

import (
	"fmt"
	"github.com/begizi/vch-server/gcp"
	"github.com/begizi/vch-server/luis"
	"github.com/begizi/vch-server/tunnel"
	"golang.org/x/net/context"
)

type Service interface {
	Voice(ctx context.Context, voice VoiceRequest) (*VoiceResponse, error)
}

func NewBasicService(client *gcp.GCPSpeechConv, queue tunnel.Queue, luis *luis.Client) Service {
	return &basicService{
		queue: queue,
		gcp:   client,
		luis:  luis,
	}
}

type basicService struct {
	queue tunnel.Queue
	gcp   *gcp.GCPSpeechConv
	luis  *luis.Client
}

func processMissingEntities(intents []*luis.CompositeEntity) []*luis.CompositeEntity {
	return intents
}

func (s basicService) Voice(_ context.Context, voice VoiceRequest) (*VoiceResponse, error) {
	transcript, err := s.gcp.Convert(voice.Audio, voice.SampleCount)
	if err != nil {
		return nil, err
	}

	resp, err := s.luis.Parse(transcript)
	if err != nil {
		return nil, fmt.Errorf("Luis Error: %v", err)
	}

	// Broadcast message with the data
	err = s.queue.Broadcast(&tunnel.QueueMessage{
		NLPResponse: tunnel.NLPResponse{
			Intents: processMissingEntities(resp.CompositeEntities),
		},
	})
	if err != nil {
		return nil, err
	}
	return &VoiceResponse{200, resp.CompositeEntities}, nil
}
