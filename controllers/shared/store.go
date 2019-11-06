package shared

import (
	"github.com/szoio/resource-operator-factory/reconciler"
	"math/rand"
	"time"
)

type Store struct {
	store map[string]reconciler.VerifyResult
}

func CreateStore() *Store {
	return &Store{store: map[string]reconciler.VerifyResult{}}
}

func (s *Store) AsyncUpdate(data string, newState reconciler.VerifyResult, d time.Duration) {
	time.Sleep(d)
	s.store[data] = newState
}

func (s *Store) Create(data string) {
	s.store[data] = reconciler.VerifyResultProvisioning
	go s.AsyncUpdate(data, reconciler.VerifyResultReady, time.Duration(rand.Intn(3)+2))
}

func (s *Store) Delete(data string) {
	s.store[data] = reconciler.VerifyResultDeleting
	go s.AsyncUpdate(data, reconciler.VerifyResultMissing, time.Duration(rand.Intn(3)+2))
}

func (s *Store) Get(data string) reconciler.VerifyResult {
	r := s.store[data]
	if r == "" {
		r = reconciler.VerifyResultMissing
	}
	return r
}
