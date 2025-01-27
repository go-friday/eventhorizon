// Copyright (c) 2014 - The Event Horizon authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"
	"github.com/looplab/eventhorizon/mocks"
)

func Test_NewAggregateStore(t *testing.T) {
	eventStore := &mocks.EventStore{
		Events: make([]eh.Event, 0),
	}
	bus := &mocks.EventBus{
		Events: make([]eh.Event, 0),
	}

	store, err := events.NewAggregateStore(nil, bus)
	if err != events.ErrInvalidEventStore {
		t.Error("there should be a ErrInvalidEventStore error:", err)
	}
	if store != nil {
		t.Error("there should be no aggregate store:", store)
	}

	store, err = events.NewAggregateStore(eventStore, nil)
	if err != events.ErrInvalidEventBus {
		t.Error("there should be a ErrInvalidEventBus error:", err)
	}
	if store != nil {
		t.Error("there should be no aggregate store:", store)
	}

	store, err = events.NewAggregateStore(eventStore, bus)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	if store == nil {
		t.Error("there should be a aggregate store")
	}
}

func Test_AggregateStore_LoadNoEvents(t *testing.T) {
	store, _, _ := createStore(t)

	ctx := context.Background()

	id := uuid.New().String()
	agg, err := store.Load(ctx, TestAggregateType, id)
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	a, ok := agg.(events.Aggregate)
	if !ok {
		t.Fatal("the aggregate shoud be of correct type")
	}
	if a.EntityID() != id {
		t.Error("the aggregate ID should be correct: ", a.EntityID(), id)
	}
	if a.Version() != 0 {
		t.Error("the version should be 0:", a.Version())
	}
}

func Test_AggregateStore_LoadEvents(t *testing.T) {
	store, eventStore, _ := createStore(t)

	ctx := context.Background()

	id := uuid.New().String()
	agg := NewTestAggregate(id)
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event1 := agg.StoreEvent(mocks.EventType, &mocks.EventData{Content: "event1"}, timestamp)
	if err := eventStore.Save(ctx, []eh.Event{event1}, 0); err != nil {
		t.Fatal("there should be no error:", err)
	}
	t.Log(eventStore.Events)

	loaded, err := store.Load(ctx, TestAggregateType, id)
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	a, ok := loaded.(events.Aggregate)
	if !ok {
		t.Fatal("the aggregate shoud be of correct type")
	}
	if a.EntityID() != id {
		t.Error("the aggregate ID should be correct: ", a.EntityID(), id)
	}
	if a.Version() != 1 {
		t.Error("the version should be 1:", a.Version())
	}
	if !reflect.DeepEqual(a.(*TestAggregate).event, event1) {
		t.Error("the event should be correct:", a.(*TestAggregate).event)
	}

	// Store error.
	eventStore.Err = errors.New("error")
	_, err = store.Load(ctx, TestAggregateType, id)
	if err == nil || err.Error() != "error" {
		t.Error("there should be an error named 'error':", err)
	}
	eventStore.Err = nil
}

func Test_AggregateStore_LoadEvents_MismatchedEventType(t *testing.T) {
	store, eventStore, _ := createStore(t)

	ctx := context.Background()

	id := uuid.New().String()
	agg := NewTestAggregate(id)
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event1 := agg.StoreEvent(mocks.EventType, &mocks.EventData{Content: "event"}, timestamp)
	if err := eventStore.Save(ctx, []eh.Event{event1}, 0); err != nil {
		t.Fatal("there should be no error:", err)
	}

	otherAggregateID := uuid.New().String()
	otherAgg := NewTestAggregateOther(otherAggregateID)
	event2 := otherAgg.StoreEvent(mocks.EventOtherType, &mocks.EventData{Content: "event2"}, timestamp)
	if err := eventStore.Save(ctx, []eh.Event{event2}, 0); err != nil {
		t.Fatal("there should be no error:", err)
	}

	loaded, err := store.Load(ctx, TestAggregateType, otherAggregateID)
	if err != events.ErrMismatchedEventType {
		t.Fatal("there should be a ErrMismatchedEventType error:", err)
	}
	if loaded != nil {
		t.Error("the aggregate should be nil")
	}
}

func Test_AggregateStore_SaveEvents(t *testing.T) {
	store, eventStore, bus := createStore(t)

	ctx := context.Background()

	id := uuid.New().String()
	agg := NewTestAggregateOther(id)
	err := store.Save(ctx, agg)
	if err != nil {
		t.Error("there should be no error:", err)
	}

	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event1 := agg.StoreEvent(mocks.EventType, &mocks.EventData{Content: "event"}, timestamp)
	err = store.Save(ctx, agg)
	if err != nil {
		t.Error("there should be no error:", err)
	}

	evts, err := eventStore.Load(ctx, id)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	if len(evts) != 1 {
		t.Fatal("there should be one event stored:", len(evts))
	}
	if evts[0] != event1 {
		t.Error("the stored event should be correct:", evts[0])
	}
	if len(agg.Events()) != 0 {
		t.Error("there should be no uncommitted events:", agg.Events())
	}
	if agg.Version() != 1 {
		t.Error("the version should be 1:", agg.Version())
	}

	if !reflect.DeepEqual(bus.Events, []eh.Event{event1}) {
		t.Error("there should be an event on the bus:", bus.Events)
	}

	// Store error.
	event1 = agg.StoreEvent(mocks.EventType, &mocks.EventData{Content: "event"}, timestamp)
	eventStore.Err = errors.New("aggregate error")
	err = store.Save(ctx, agg)
	if err == nil || err.Error() != "aggregate error" {
		t.Error("there should be an error named 'error':", err)
	}
	eventStore.Err = nil

	// Aggregate error.
	event1 = agg.StoreEvent(mocks.EventType, &mocks.EventData{Content: "event"}, timestamp)
	agg.err = errors.New("error")
	err = store.Save(ctx, agg)
	if _, ok := err.(events.ApplyEventError); !ok {
		t.Error("there should be an error of type ApplyEventError:", err)
	}
	agg.err = nil

	// Bus error.
	event1 = agg.StoreEvent(mocks.EventType, &mocks.EventData{Content: "event"}, timestamp)
	bus.Err = errors.New("bus error")
	err = store.Save(ctx, agg)
	if err == nil || err.Error() != "bus error" {
		t.Error("there should be an error named 'error':", err)
	}
}

func Test_AggregateStore_AggregateNotRegistered(t *testing.T) {
	store, _, _ := createStore(t)

	ctx := context.Background()

	id := uuid.New().String()
	agg, err := store.Load(ctx, "TestAggregate3", id)
	if err != eh.ErrAggregateNotRegistered {
		t.Error("there should be a eventhorizon.ErrAggregateNotRegistered error:", err)
	}
	if agg != nil {
		t.Fatal("there should be no aggregate")
	}
}

func createStore(t *testing.T) (*events.AggregateStore, *mocks.EventStore, *mocks.EventBus) {
	eventStore := &mocks.EventStore{
		Events: make([]eh.Event, 0),
	}
	bus := &mocks.EventBus{
		Events: make([]eh.Event, 0),
	}
	store, err := events.NewAggregateStore(eventStore, bus)
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	if store == nil {
		t.Fatal("there should be a aggregate store")
	}
	return store, eventStore, bus
}

func init() {
	eh.RegisterAggregate(func(id eh.ID) eh.Aggregate {
		return NewTestAggregateOther(id)
	})
}

const TestAggregateOtherType eh.AggregateType = "TestAggregateOther"

type TestAggregateOther struct {
	*events.AggregateBase
	err error
}

var _ = events.Aggregate(&TestAggregateOther{})

func NewTestAggregateOther(id eh.ID) *TestAggregateOther {
	return &TestAggregateOther{
		AggregateBase: events.NewAggregateBase(TestAggregateOtherType, id),
	}
}

func (a *TestAggregateOther) HandleCommand(ctx context.Context, cmd eh.Command) error {
	return nil
}

func (a *TestAggregateOther) ApplyEvent(ctx context.Context, event eh.Event) error {
	if a.err != nil {
		return a.err
	}
	return nil
}
