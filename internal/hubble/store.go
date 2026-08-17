package hubble

import "sync"

type Store struct {
	mu    sync.Mutex
	flows map[string][]FlowEvent
	limit int
}

func NewStore(limitPerSession int) *Store {
	return &Store{
		flows: make(map[string][]FlowEvent),
		limit: limitPerSession,
	}
}

func (s *Store) Add(event FlowEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.flows[event.Session]
	list = append(list, event)

	if len(list) > s.limit {
		list = list[len(list)-s.limit:]
	}

	s.flows[event.Session] = list
}

func (s *Store) Get(session string) []FlowEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]FlowEvent{}, s.flows[session]...)
}
