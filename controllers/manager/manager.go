package manager

import (
	"fmt"
	"github.com/szoio/resource-operator-factory/reconciler"
	"math/rand"
	"time"
)

type Event string

const (
	EventCreate Event = "Create"
	EventGet    Event = "Get"
	EventUpdate Event = "Update"
	EventDelete Event = "Delete"
)

type Data struct {
	States     []reconciler.VerifyResult
	Events     []Event
	Behaviours []Behaviour
}

func (sd *Data) Set(r reconciler.VerifyResult) {
	sd.States = append(sd.States, r)
}

type Manager struct {
	dataStore map[string]*Data
}

func (m *Manager) Set(id string, r reconciler.VerifyResult) {
	x := m.getOrCreate(id)
	x.States = append(x.States, r)
}

func (m *Manager) addEvent(id string, event Event) {
	x := m.getOrCreate(id)
	x.Events = append(x.Events, event)
}

func (m *Manager) getOrCreate(id string) *Data {
	x := m.dataStore[id]
	if x == nil {
		x = &Data{
			States: []reconciler.VerifyResult{},
			Events: []Event{},
		}
		m.dataStore[id] = x
	}
	return x
}

func CreateManager() *Manager {
	return &Manager{dataStore: map[string]*Data{}}
}

func (m *Manager) Create(id string) (reconciler.ApplyResult, error) {
	result, err := m.apply(id, EventCreate)
	return reconciler.ApplyResult(result), err
}

func (m *Manager) Update(id string) (reconciler.ApplyResult, error) {
	result, err := m.apply(id, EventUpdate)
	return reconciler.ApplyResult(result), err
}

func (m *Manager) Delete(id string) (reconciler.DeleteResult, error) {
	result, err := m.apply(id, EventDelete)
	return reconciler.DeleteResult(result), err
}

func (m *Manager) Get(id string) (reconciler.VerifyResult, error) {
	result, err := m.apply(id, EventGet)
	return reconciler.VerifyResult(result), err
}

func (m *Manager) Clear(id string) {
	for k := range m.dataStore {
		delete(m.dataStore, k)
	}
}

func (m *Manager) AddBehaviour(id string, b Behaviour) {
	x := m.getOrCreate(id)
	x.Behaviours = append(x.Behaviours, b)
}

func (m *Manager) ClearBehaviours(id string) {
	x := m.getOrCreate(id)
	x.Behaviours = []Behaviour{}
}

func (m *Manager) GetRecord(id string) *Data {
	r := m.dataStore[id]
	if r == nil {
		r = &Data{}
	}
	return r
}

func (m *Manager) asyncUpdate(id string, newState reconciler.VerifyResult, d time.Duration) {
	time.Sleep(d)
	m.Set(id, newState)
}

func (m *Manager) apply(id string, event Event) (string, error) {
	operation := m.getOperation(id, event)
	m.addEvent(id, event)
	return operation(m, id)
}

func (m *Manager) getOperation(id string, event Event) Operation {
	x := m.getOrCreate(id)
	// count the number of events of type Event
	eventCount := m.countEvents(x, event)
	var behaviour *Behaviour = nil
	for _, b := range x.Behaviours {
		if b.Event == event && b.From <= eventCount && (b.Count == 0 || b.From+b.Count > eventCount) {
			behaviour = &b
		}
	}
	if behaviour != nil {
		return behaviour.Operation
	}
	switch event {
	case EventCreate:
		return CreateAsync.AsOperation()
	case EventUpdate:
		return UpdateSync.AsOperation()
	case EventGet:
		return GetStandard.AsOperation()
	case EventDelete:
		return DeleteAsync.AsOperation()
	}
	return nil
}

func (m *Manager) CountEvents(id string, event Event) int {
	return m.countEvents(m.getOrCreate(id), event)
}

func (m *Manager) countEvents(x *Data, event Event) int {
	count := 0
	for _, e := range x.Events {
		if e == event {
			count++
		}
	}
	return count
}

type Behaviour struct {
	Event     Event
	Operation Operation
	From      int
	Count     int
}

func (x ApplyOperation) AsOperation() Operation {
	return func(m *Manager, id string) (s string, e error) {
		r, e := x(m, id)
		return string(r), e
	}
}

func (x DeleteOperation) AsOperation() Operation {
	return func(m *Manager, id string) (s string, e error) {
		r, e := x(m, id)
		return string(r), e
	}
}

func (x GetOperation) AsOperation() Operation {
	return func(m *Manager, id string) (s string, e error) {
		r, e := x(m, id)
		return string(r), e
	}
}

const (
	startMillis = 400
	endMillis   = 500
)

func randomDelay(startMillis int, endMillis int) time.Duration {
	delay := time.Duration(rand.Intn(startMillis)+(endMillis-startMillis)) * time.Millisecond
	return delay
}

type Operation func(m *Manager, id string) (string, error)
type ApplyOperation func(m *Manager, id string) (reconciler.ApplyResult, error)
type GetOperation func(m *Manager, id string) (reconciler.VerifyResult, error)
type DeleteOperation func(m *Manager, id string) (reconciler.DeleteResult, error)

// Include operations below
var CreateAsync ApplyOperation = func(m *Manager, id string) (reconciler.ApplyResult, error) {
	m.Set(id, reconciler.VerifyResultProvisioning)
	go m.asyncUpdate(id, reconciler.VerifyResultReady, randomDelay(startMillis, endMillis))
	return reconciler.ApplyResultAwaitingVerification, nil
}

var CreateSync ApplyOperation = func(m *Manager, id string) (reconciler.ApplyResult, error) {
	m.Set(id, reconciler.VerifyResultReady)
	return reconciler.ApplyResultSucceeded, nil
}

var CreateFail ApplyOperation = func(m *Manager, id string) (reconciler.ApplyResult, error) {
	m.Set(id, reconciler.VerifyResultError)
	return reconciler.ApplyResultError, fmt.Errorf("error creating resource")
}

var UpdateSync = CreateSync
var UpdateAsync = CreateAsync

var DeleteAsync DeleteOperation = func(m *Manager, id string) (reconciler.DeleteResult, error) {
	m.Set(id, reconciler.VerifyResultDeleting)
	go m.asyncUpdate(id, reconciler.VerifyResultMissing, randomDelay(startMillis, endMillis))
	return reconciler.DeleteAwaitingVerification, nil
}

var DeleteSync DeleteOperation = func(m *Manager, id string) (reconciler.DeleteResult, error) {
	m.Set(id, reconciler.VerifyResultMissing)
	return reconciler.DeleteSucceeded, nil
}

var GetStandard GetOperation = func(m *Manager, id string) (reconciler.VerifyResult, error) {
	x := m.GetRecord(id)
	if x.States == nil || len(x.States) == 0 {
		return reconciler.VerifyResultMissing, nil
	}
	return x.States[len(x.States)-1], nil
}

var VerifyFail GetOperation = func(m *Manager, id string) (reconciler.VerifyResult, error) {
	m.Set(id, reconciler.VerifyResultError)
	return reconciler.VerifyResultError, fmt.Errorf("failed to verify resource")
}

var VerifyNeedsUpdate GetOperation = func(m *Manager, id string) (reconciler.VerifyResult, error) {
	m.Set(id, reconciler.VerifyResultUpdateRequired)
	return reconciler.VerifyResultUpdateRequired, nil
}

var CreateCompleteFail ApplyOperation = func(m *Manager, id string) (reconciler.ApplyResult, error) {
	m.Set(id, reconciler.VerifyResultProvisioning)
	go m.asyncUpdate(id, reconciler.VerifyResultError, randomDelay(startMillis, endMillis))
	return reconciler.ApplyResultAwaitingVerification, nil
}
