package fspider

import "sync"

type PathSet interface {
	Add(string)
	Remove(string)
	Contains(string) bool
	Paths() []string
}

// 并发安全的字符串集合
type pathSetImpl struct {
	m map[string]struct{}
	l *sync.RWMutex
}

func NewPathSet() *pathSetImpl {
	return &pathSetImpl{
		m: make(map[string]struct{}),
		l: &sync.RWMutex{},
	}
}
func (s *pathSetImpl) Add(str string) {
	s.l.Lock()
	s.m[str] = struct{}{}
	s.l.Unlock()
}
func (s *pathSetImpl) Remove(str string) {
	s.l.Lock()
	delete(s.m, str)
	s.l.Unlock()
}
func (s *pathSetImpl) Contains(str string) bool {
	s.l.RLock()
	_, ok := s.m[str]
	s.l.RUnlock()
	return ok
}

func (s *pathSetImpl) Paths() []string {
	s.l.RLock()
	defer s.l.RUnlock()
	ret := make([]string, 0, len(s.m))
	for k := range s.m {
		ret = append(ret, k)
	}
	return ret
}
