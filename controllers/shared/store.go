package shared

import (
	"github.com/szoio/resource-operator-factory/reconciler"
	"math/rand"
	"time"
)

type StoreData struct {
	States []reconciler.VerifyResult
	Events []string
}

func (sd *StoreData) Set(r reconciler.VerifyResult) {
	sd.States = append(sd.States, r)
}

type Store struct {
	store map[string]*StoreData
}

func (s *Store) Set(id string, r reconciler.VerifyResult) {
	x := s.getOrCreate(id)
	x.States = append(x.States, r)
}

func (s *Store) AddEvent(id string, event string) {
	x := s.getOrCreate(id)
	x.Events = append(x.Events, event)
}

func (s *Store) getOrCreate(id string) *StoreData {
	x := s.store[id]
	if x == nil {
		x = &StoreData{
			States: []reconciler.VerifyResult{},
			Events: []string{},
		}
		s.store[id] = x
	}
	return x
}

func CreateStore() *Store {
	return &Store{store: map[string]*StoreData{}}
}

func (s *Store) AsyncUpdate(id string, newState reconciler.VerifyResult, d time.Duration) {
	time.Sleep(d)
	s.Set(id, newState)
}

func (s *Store) Create(id string) {
	s.AddEvent(id, "Create")
	s.Set(id, reconciler.VerifyResultProvisioning)
	go s.AsyncUpdate(id, reconciler.VerifyResultReady, time.Duration(rand.Intn(3)+2))
}

func (s *Store) Delete(id string) {
	s.AddEvent(id, "Delete")
	s.Set(id, reconciler.VerifyResultDeleting)
	go s.AsyncUpdate(id, reconciler.VerifyResultMissing, time.Duration(rand.Intn(3)+2))
}

func (s *Store) GetRecord(id string) *StoreData {
	r := s.store[id]
	if r == nil {
		r = &StoreData{}
	}
	return r
}

func (s *Store) Get(id string) reconciler.VerifyResult {
	s.AddEvent(id, "Get")
	x := s.GetRecord(id)
	if x.States == nil || len(x.States) == 0 {
		return reconciler.VerifyResultMissing
	}
	return x.States[len(x.States) - 1]
}

func (s *Store) Clear(id string) {
	for k := range s.store {
		delete(s.store, k)
	}
}

