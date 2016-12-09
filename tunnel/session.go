package tunnel

import (
	"errors"
	"sync"
	"time"

	"github.com/begizi/vch-server/pb"
)

type SessionId string

type Session struct {
	Id     SessionId
	Start  time.Time
	Stream pb.VCH_TunnelServer
}

type SessionStore struct {
	mtx      sync.RWMutex
	sessions map[SessionId]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[SessionId]*Session),
	}
}

func (s *SessionStore) Add(session *Session) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.sessions[session.Id] = session
	return nil
}

func (s *SessionStore) Remove(id SessionId) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, ok := s.sessions[id]
	if ok {
		delete(s.sessions, id)
		return nil
	}
	return errors.New("Not Found")
}

func (s *SessionStore) FindBySessionId(id SessionId) (*Session, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	session, ok := s.sessions[id]
	if ok {
		return session, nil
	}
	return nil, errors.New("Not Found")
}

func (s *SessionStore) List() ([]*Session, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	sessions := []*Session{}
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}
