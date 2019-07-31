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

// Package eventhorizon is a CQRS/ES toolkit.
package eventhorizon

// ID is identify of an Entity
//
// Data type should be in type of string.
// For example, Netflix use "S1:C1" as ID for entity, meaning "Session 1 - Chapter 1"
type ID = string

// Entity is an item which is identified by an ID.
//
// From http://cqrs.nu/Faq:
// "Entities or reference types are characterized by having an identity that's
// not tied to their attribute values. All attributes in an entity can change
// and it's still "the same" entity. Conversely, two entities might be
// equivalent in all their attributes, but will still be distinct."
type Entity interface {
	// EntityID returns the ID of the entity.
	EntityID() ID
}

// NilID is pre-defined nil value for ID
// when you need to returns from function.
const NilID ID = ""

// IsNilID check a ID is nil or not.
//
// ID is an empty string or UUID format with all zeros.
func IsNilID(id ID) bool {
	return id == NilID || id == "00000000-0000-0000-0000-000000000000"
}
