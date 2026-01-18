package internal

type Store struct {
	m map[string]string
}

func NewStore() *Store {
	return &Store{m: make(map[string]string)}
}

func (s *Store) Set(key string, value string) error {
	s.m[key] = value
	return nil
}

func (s *Store) Get(key string) (string, bool, error) {
	if value, ok := s.m[key]; !ok {
		return "", false, nil
	} else {
		return value, true, nil
	}
}
