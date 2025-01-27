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

package aggregate_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/aggregate"
	"github.com/looplab/eventhorizon/mocks"
)

func Test_NewCommandHandler(t *testing.T) {
	store := &mocks.AggregateStore{
		Aggregates: make(map[eh.ID]eh.Aggregate),
	}
	h, err := aggregate.NewCommandHandler(mocks.AggregateType, store)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	if h == nil {
		t.Error("there should be a handler")
	}

	h, err = aggregate.NewCommandHandler(mocks.AggregateType, nil)
	if err != aggregate.ErrNilAggregateStore {
		t.Error("there should be a ErrNilAggregateStore error:", err)
	}
	if h != nil {
		t.Error("there should be no handler:", h)
	}
}

func Test_CommandHandler(t *testing.T) {
	a, h, _ := createAggregateAndHandler(t)

	ctx := context.WithValue(context.Background(), "testkey", "testval")

	cmd := &mocks.Command{
		ID:      a.EntityID(),
		Content: "command1",
	}
	err := h.HandleCommand(ctx, cmd)
	if err != nil {
		t.Error("there should be no error:", err)
	}
	if !reflect.DeepEqual(a.Commands, []eh.Command{cmd}) {
		t.Error("the handeled command should be correct:", a.Commands)
	}
	if val, ok := a.Context.Value("testkey").(string); !ok || val != "testval" {
		t.Error("the context should be correct:", a.Context)
	}
}

func Test_CommandHandler_AggregateNotFound(t *testing.T) {
	store := &mocks.AggregateStore{
		Aggregates: map[eh.ID]eh.Aggregate{},
	}
	h, err := aggregate.NewCommandHandler(mocks.AggregateType, store)
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	if h == nil {
		t.Fatal("there should be a handler")
	}

	cmd := &mocks.Command{
		ID:      uuid.New().String(),
		Content: "command1",
	}
	err = h.HandleCommand(context.Background(), cmd)
	if err != eh.ErrAggregateNotFound {
		t.Error("there should be a command error:", err)
	}
}

func Test_CommandHandler_ErrorInHandler(t *testing.T) {
	a, h, _ := createAggregateAndHandler(t)

	a.Err = errors.New("command error")
	cmd := &mocks.Command{
		ID:      a.EntityID(),
		Content: "command1",
	}
	err := h.HandleCommand(context.Background(), cmd)
	if err == nil || err.Error() != "command error" {
		t.Error("there should be a command error:", err)
	}
	if !reflect.DeepEqual(a.Commands, []eh.Command{}) {
		t.Error("the handeled command should be correct:", a.Commands)
	}
}

func Test_CommandHandler_ErrorWhenSaving(t *testing.T) {
	a, h, store := createAggregateAndHandler(t)

	store.Err = errors.New("save error")
	cmd := &mocks.Command{
		ID:      a.EntityID(),
		Content: "command1",
	}
	err := h.HandleCommand(context.Background(), cmd)
	if err == nil || err.Error() != "save error" {
		t.Error("there should be a command error:", err)
	}
}

func Test_CommandHandler_NoHandlers(t *testing.T) {
	_, h, _ := createAggregateAndHandler(t)

	cmd := &mocks.Command{
		ID:      uuid.New().String(),
		Content: "command1",
	}
	err := h.HandleCommand(context.Background(), cmd)
	if err != eh.ErrAggregateNotFound {
		t.Error("there should be a ErrAggregateNotFound error:", nil)
	}
}

func BenchmarkCommandHandler(b *testing.B) {
	a := mocks.NewAggregate(uuid.New().String())
	store := &mocks.AggregateStore{
		Aggregates: map[eh.ID]eh.Aggregate{
			a.EntityID(): a,
		},
	}
	h, err := aggregate.NewCommandHandler(mocks.AggregateType, store)
	if err != nil {
		b.Fatal("there should be no error:", err)
	}

	ctx := context.WithValue(context.Background(), "testkey", "testval")

	cmd := &mocks.Command{
		ID:      a.EntityID(),
		Content: "command1",
	}
	for i := 0; i < b.N; i++ {
		h.HandleCommand(ctx, cmd)
	}
	if len(a.Commands) != b.N {
		b.Error("the num handled commands should be correct:", len(a.Commands), b.N)
	}
}

func createAggregateAndHandler(t *testing.T) (*mocks.Aggregate, *aggregate.CommandHandler, *mocks.AggregateStore) {
	a := mocks.NewAggregate(uuid.New().String())
	store := &mocks.AggregateStore{
		Aggregates: map[eh.ID]eh.Aggregate{
			a.EntityID(): a,
		},
	}
	h, err := aggregate.NewCommandHandler(mocks.AggregateType, store)
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	return a, h, store
}
