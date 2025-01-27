// Copyright (c) 2017 - The Event Horizon authors.
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

package projector_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/eventhandler/projector"
	"github.com/looplab/eventhorizon/mocks"
	"github.com/looplab/eventhorizon/repo/version"
)

func Test_EventHandler_CreateModel(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.SimpleModel{}
	})

	ctx := context.Background()

	// Driver creates entity.
	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	entity := &mocks.SimpleModel{
		ID: id,
	}
	repo.LoadErr = eh.RepoError{
		Err: eh.ErrEntityNotFound,
	}
	proj.newEntity = entity
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Error("there shoud be no error:", err)
	}
	if proj.event != event {
		t.Error("the handled event should be correct:", proj.event)
	}
	if !reflect.DeepEqual(proj.entity, &mocks.SimpleModel{}) {
		t.Error("the entity should be correct:", proj.entity)
	}
	if repo.Entity != proj.newEntity {
		t.Error("the new entity should be correct:", repo.Entity)
	}
}

func Test_EventHandler_UpdateModel(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.SimpleModel{}
	})

	ctx := context.Background()

	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	entity := &mocks.SimpleModel{
		ID: id,
	}
	repo.Entity = entity
	proj.newEntity = &mocks.SimpleModel{
		ID:      id,
		Content: "updated",
	}
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Error("there shoud be no error:", err)
	}
	if proj.event != event {
		t.Error("the handled event should be correct:", proj.event)
	}
	if proj.entity != entity {
		t.Error("the entity should be correct:", proj.entity)
	}
	if repo.Entity != proj.newEntity {
		t.Error("the new entity should be correct:", repo.Entity)
	}
}

func Test_EventHandler_UpdateModelWithVersion(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.Model{}
	})

	ctx := context.Background()

	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	entity := &mocks.Model{
		ID: id,
	}
	repo.Entity = entity
	proj.newEntity = &mocks.Model{
		ID:      id,
		Version: 1,
		Content: "version 1",
	}
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Error("there shoud be no error:", err)
	}
	if proj.event != event {
		t.Error("the handled event should be correct:", proj.event)
	}
	if proj.entity != entity {
		t.Error("the entity should be correct:", proj.entity)
	}
	if repo.Entity != proj.newEntity {
		t.Error("the new entity should be correct:", repo.Entity)
	}
}

func Test_EventHandler_UpdateModelWithEventsOutOfOrder(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, version.NewRepo(repo))
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.Model{}
	})

	ctx := context.Background()

	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 3)
	entity := &mocks.Model{
		ID:      id,
		Version: 1,
		Content: "version 1",
	}
	newEntity := &mocks.Model{
		ID:      id,
		Version: 2,
		Content: "version 2",
	}
	repo.Entity = entity
	proj.newEntity = &mocks.Model{
		ID:      id,
		Version: 3,
		Content: "version 3",
	}
	go func() {
		<-time.After(100 * time.Millisecond)
		repo.Entity = newEntity
	}()
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Error("there shoud be no error:", err)
	}
	if proj.event != event {
		t.Error("the handled event should be correct:", proj.event)
	}
	if proj.entity != newEntity {
		t.Error("the entity should be correct:", proj.entity)
	}
	if repo.Entity != proj.newEntity {
		t.Error("the new entity should be correct:", repo.Entity)
	}
}

func Test_EventHandler_DeleteModel(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.SimpleModel{}
	})

	ctx := context.Background()

	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	entity := &mocks.SimpleModel{
		ID: id,
	}
	repo.Entity = entity
	proj.newEntity = nil
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Error("there shoud be no error:", err)
	}
	if proj.event != event {
		t.Error("the handled event should be correct:", proj.event)
	}
	if proj.entity != entity {
		t.Error("the entity should be correct:", proj.entity)
	}
	if repo.Entity != proj.newEntity {
		t.Error("the new entity should be correct:", repo.Entity)
	}
}

func Test_EventHandler_LoadError(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.SimpleModel{}
	})

	ctx := context.Background()

	// Driver creates entity.
	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	loadErr := errors.New("load error")
	repo.LoadErr = loadErr
	expectedErr := projector.Error{
		Err:       loadErr,
		Namespace: eh.NamespaceFromContext(ctx),
	}
	if err := handler.HandleEvent(ctx, event); !reflect.DeepEqual(err, expectedErr) {
		t.Error("there shoud be an error:", err)
	}
}

func Test_EventHandler_SaveError(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.SimpleModel{}
	})

	ctx := context.Background()

	// Driver creates entity.
	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	saveErr := errors.New("save error")
	repo.SaveErr = saveErr
	expectedErr := projector.Error{
		Err:       saveErr,
		Namespace: eh.NamespaceFromContext(ctx),
	}
	if err := handler.HandleEvent(ctx, event); !reflect.DeepEqual(err, expectedErr) {
		t.Error("there shoud be an error:", err)
	}
}

func Test_EventHandler_ProjectError(t *testing.T) {
	repo := &mocks.Repo{}
	proj := &TestProjector{}
	handler := projector.NewEventHandler(proj, repo)
	handler.SetEntityFactory(func() eh.Entity {
		return &mocks.SimpleModel{}
	})

	ctx := context.Background()

	// Driver creates entity.
	id := uuid.New().String()
	eventData := &mocks.EventData{Content: "event1"}
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	event := eh.NewEventForAggregate(mocks.EventType, eventData, timestamp,
		mocks.AggregateType, id, 1)
	projectErr := errors.New("save error")
	proj.err = projectErr
	expectedErr := projector.Error{
		Err:       projectErr,
		Namespace: eh.NamespaceFromContext(ctx),
	}
	if err := handler.HandleEvent(ctx, event); !reflect.DeepEqual(err, expectedErr) {
		t.Error("there shoud be an error:", err)
	}
}

const (
	TestProjectorType projector.Type = "TestProjector"
)

type TestProjector struct {
	event             eh.Event
	context           context.Context
	entity, newEntity eh.Entity
	// Used to simulate errors in the store.
	err error
}

func (m *TestProjector) ProjectorType() projector.Type {
	return TestProjectorType
}

func (m *TestProjector) Project(ctx context.Context, event eh.Event, entity eh.Entity) (eh.Entity, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.context = ctx
	m.event = event
	m.entity = entity
	return m.newEntity, nil
}
