// Copyright (c) 2016 - The Event Horizon authors.
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

package eventhorizon_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
)

func Test_NewEvent(t *testing.T) {
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEvent(TestEventType, &TestEventData{"event1"}, timestamp)
	if event.EventType() != TestEventType {
		t.Error("the event type should be correct:", event.EventType())
	}
	if !reflect.DeepEqual(event.Data(), &TestEventData{"event1"}) {
		t.Error("the data should be correct:", event.Data())
	}
	if !event.Timestamp().Equal(timestamp) {
		t.Error("the timestamp should not be zero:", event.Timestamp())
	}
	if event.Version() != 0 {
		t.Error("the version should be zero:", event.Version())
	}
	if event.String() != "TestEvent@0" {
		t.Error("the string representation should be correct:", event.String())
	}

	id := uuid.New().String()
	event = eh.NewEventForAggregate(TestEventType, &TestEventData{"event1"}, timestamp,
		TestAggregateType, id, 3)
	if event.EventType() != TestEventType {
		t.Error("the event type should be correct:", event.EventType())
	}
	if !reflect.DeepEqual(event.Data(), &TestEventData{"event1"}) {
		t.Error("the data should be correct:", event.Data())
	}
	if !event.Timestamp().Equal(timestamp) {
		t.Error("the timestamp should not be zero:", event.Timestamp())
	}
	if event.AggregateType() != TestAggregateType {
		t.Error("the aggregate type should be correct:", event.AggregateType())
	}
	if event.AggregateID() != id {
		t.Error("the aggregate ID should be correct:", event.AggregateID())
	}
	if event.Version() != 3 {
		t.Error("the version should be zero:", event.Version())
	}
	if event.String() != "TestEvent@3" {
		t.Error("the string representation should be correct:", event.String())
	}
}

func Test_CreateEventData(t *testing.T) {
	data, err := eh.CreateEventData(TestEventRegisterType)
	if err != eh.ErrEventDataNotRegistered {
		t.Error("there should be a event not registered error:", err)
	}

	eh.RegisterEventData(TestEventRegisterType, func() eh.EventData {
		return &TestEventRegisterData{}
	})

	data, err = eh.CreateEventData(TestEventRegisterType)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	if _, ok := data.(*TestEventRegisterData); !ok {
		t.Errorf("the event type should be correct: %T", data)
	}

	eh.UnregisterEventData(TestEventRegisterType)
}

func Test_RegisterEventEmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || r != "eventhorizon: attempt to register empty event type" {
			t.Error("there should have been a panic:", r)
		}
	}()
	eh.RegisterEventData(TestEventRegisterEmptyType, func() eh.EventData {
		return &TestEventRegisterEmptyData{}
	})
}

func Test_RegisterEventTwice(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || r != "eventhorizon: registering duplicate types for \"TestEventRegisterTwice\"" {
			t.Error("there should have been a panic:", r)
		}
	}()
	eh.RegisterEventData(TestEventRegisterTwiceType, func() eh.EventData {
		return &TestEventRegisterTwiceData{}
	})
	eh.RegisterEventData(TestEventRegisterTwiceType, func() eh.EventData {
		return &TestEventRegisterTwiceData{}
	})
}

func Test_UnregisterEventEmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || r != "eventhorizon: attempt to unregister empty event type" {
			t.Error("there should have been a panic:", r)
		}
	}()
	eh.UnregisterEventData(TestEventUnregisterEmptyType)
}

func Test_UnregisterEventTwice(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || r != "eventhorizon: unregister of non-registered type \"TestEventUnregisterTwice\"" {
			t.Error("there should have been a panic:", r)
		}
	}()
	eh.RegisterEventData(TestEventUnregisterTwiceType, func() eh.EventData {
		return &TestEventUnregisterTwiceData{}
	})
	eh.UnregisterEventData(TestEventUnregisterTwiceType)
	eh.UnregisterEventData(TestEventUnregisterTwiceType)
}

const (
	TestEventType                eh.EventType = "TestEvent"
	TestEventRegisterType        eh.EventType = "TestEventRegister"
	TestEventRegisterEmptyType   eh.EventType = ""
	TestEventRegisterTwiceType   eh.EventType = "TestEventRegisterTwice"
	TestEventUnregisterEmptyType eh.EventType = ""
	TestEventUnregisterTwiceType eh.EventType = "TestEventUnregisterTwice"
)

type TestEventData struct {
	Content string
}

type TestEventRegisterData struct{}

type TestEventRegisterEmptyData struct{}

type TestEventRegisterTwiceData struct{}

type TestEventUnregisterTwiceData struct{}
