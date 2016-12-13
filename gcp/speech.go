package gcp

import (
	"context"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"google.golang.org/grpc"

	gcontext "golang.org/x/net/context"
	speech "google.golang.org/genproto/googleapis/cloud/speech/v1beta1"
)

type GCPSpeechConv struct {
	ctx  gcontext.Context
	conn *grpc.ClientConn

	client speech.SpeechClient
}

func NewGCPSpeechConv() (*GCPSpeechConv, error) {
	ctx := context.Background()
	conn, err := transport.DialGRPC(ctx,
		option.WithEndpoint("speech.googleapis.com:443"),
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"),
		option.WithServiceAccountFile("./credentials-key.json"),
		option.WithGRPCDialOption(grpc.WithBackoffMaxDelay(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	client := speech.NewSpeechClient(conn)

	return &GCPSpeechConv{ctx, conn, client}, nil
}

func (gcp *GCPSpeechConv) Convert(data []byte, sampleRate uint32) (string, error) {
	resp, err := gcp.recognize(data, sampleRate)
	if err != nil {
		return "", err
	}

	var best *speech.SpeechRecognitionAlternative

	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			if best == nil || alt.Confidence > best.Confidence {
				best = alt
			}
		}
	}

	if best == nil {
		return "", nil
	}

	return best.Transcript, nil
}

func (gcp *GCPSpeechConv) recognize(data []byte, sampleRate uint32) (*speech.SyncRecognizeResponse, error) {
	return gcp.client.SyncRecognize(gcp.ctx, &speech.SyncRecognizeRequest{
		Config: &speech.RecognitionConfig{
			Encoding:   speech.RecognitionConfig_LINEAR16,
			SampleRate: int32(sampleRate),
		},
		Audio: &speech.RecognitionAudio{
			AudioSource: &speech.RecognitionAudio_Content{Content: data},
		},
	})
}
