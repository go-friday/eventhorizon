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

package saga_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/eventhandler/saga"
	"github.com/looplab/eventhorizon/mocks"
)

func Test_EventHandler(t *testing.T) {
	commandHandler := &mocks.CommandHandler{
		Commands: []eh.Command{},
	}
	sg := &TestSaga{}
	handler := saga.NewEventHandler(sg, commandHandler)

	ctx := context.Background()

	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	sg.commands = []eh.Command{&mocks.Command{ID: uuid.New().String(), Content: "content"}}
	handler.HandleEvent(ctx, event)
	if sg.event != event {
		t.Error("the handled event should be correct:", sg.event)
	}
	if !reflect.DeepEqual(commandHandler.Commands, sg.commands) {
		t.Error("the produced commands should be correct:", commandHandler.Commands)
	}
}

const (
	TestSagaType saga.Type = "TestSaga"
)

type TestSaga struct {
	event    eh.Event
	context  context.Context
	commands []eh.Command
}

func (m *TestSaga) SagaType() saga.Type {
	return TestSagaType
}

func (m *TestSaga) RunSaga(ctx context.Context, event eh.Event) []eh.Command {
	m.event = event
	m.context = ctx
	return m.commands
}
