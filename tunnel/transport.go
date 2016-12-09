package tunnel

import (
	"github.com/begizi/vch-server/pb"
	"github.com/go-kit/kit/log"
	"github.com/satori/go.uuid"
	"time"
)

type VCHTunnelServer struct {
	queue Queue

	// Session Store for adding new sessions
	sessions *SessionStore

	// Message logger
	logger log.Logger
}

func (s VCHTunnelServer) SendToStream(message NLPResponse) error {
	sessions, err := s.sessions.List()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		session.Stream.Send(&pb.TunnelResponse{
			Event: &pb.TunnelResponse_Response{
				Response: &pb.NLPResponse{
					Body: message.Body,
				},
			},
		})
	}

	return nil
}

// Tunnel transport handler
func (s *VCHTunnelServer) Tunnel(req *pb.TunnelRequest, stream pb.VCH_TunnelServer) error {
	streamCtx := stream.Context()

	id := uuid.NewV4()
	newSession := &Session{
		Id:     SessionId(id.String()),
		Start:  time.Now(),
		Stream: stream,
	}

	err := s.sessions.Add(newSession)
	if err != nil {
		return err
	}
	s.logger.Log("msg", "Added stream to list", "streamId", newSession.Id)

	for {
		select {
		case <-streamCtx.Done():
			err := streamCtx.Err()
			s.logger.Log("msg", "Stream done", "sessionId", newSession.Id, "err", err)
			return s.sessions.Remove(newSession.Id)
		}

	}
}

func MakeTunnelServer(q Queue, logger log.Logger) (*VCHTunnelServer, error) {
	queuec, err := q.Listen()
	if err != nil {
		return nil, err
	}

	sessions := NewSessionStore()

	server := &VCHTunnelServer{
		logger:   logger,
		queue:    q,
		sessions: sessions,
	}

	// Process for handling queue messages
	go func() {
		for {
			select {

			// process incoming messages from RedisPubSub, and send messages.
			case msg := <-queuec:
				if msg == nil {
					logger.Log("msg", "Message Channel has closed. Exiting.")
					return
				}

				logger.Log("msg", "Sending a new message to all sessions")
				server.SendToStream(msg.NLPResponse)
			}

		}
	}()

	return server, nil
}
