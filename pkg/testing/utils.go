/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package testing

import (
	"strings"
	"reflect"
	"testing"

	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/cache"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

// GetKey is a helper function used by controllers unit tests to get the
// key for a given kubernetes resource.
func GetKey(obj interface{}, t *testing.T) string {
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// if tombstone , try getting the value from tombstone.Obj
		obj = tombstone.Obj
	}
	val := reflect.ValueOf(obj).Elem()
	name := val.FieldByName("Name").String()
	kind := val.FieldByName("Kind").String()
	// Note kind is not always set in the tests, so ignoring that for now
	if len(name) == 0 || len(kind) == 0 {
		t.Errorf("Unexpected object %v", obj)
	}

	key, err := keyFunc(obj)
	if err != nil {
		t.Errorf("Unexpected error getting key for %v %v: %v", kind, name, err)
		return ""
	}
	return key
}

// ExpectEvent will check if expected event exists under the recorder
func ExpectEvent(recorder *record.FakeRecorder, reason string, t *testing.T) {
	select {
    case e, ok := <-recorder.Events:
        if ok && strings.Contains(e, reason) {
			return
		}
		t.Errorf("Failed to find event with the reason \"%s\"", reason)
    default:
        t.Errorf("Failed to find event with the reason \"%s\"", reason)
    }
}