// Copyright 2016-present Liferay, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonlib

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

// AssertJSONMarshal asserts that a given object marshals to JSON correctly
func AssertJSONMarshal(t *testing.T, want string, got interface{}) {
	var wantJSON interface{}
	var gotMap interface{}

	bin, err := json.Marshal(got)

	if err != nil {
		t.Error(err)
	}

	if err = json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Errorf("Wanted value %s isn't JSON.", want)
	}

	if err = json.Unmarshal(bin, &gotMap); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(wantJSON, gotMap) {
		marshalError(t, wantJSON, gotMap)
	}
}

func marshalError(t *testing.T, wantJSON, gotMap interface{}) {
	var friendly string

	switch wantJSON != nil && gotMap != nil {
	case true:
		friendly = pretty.Compare(wantJSON, gotMap)
	default:
		friendly = fmt.Sprintf("%v instead of %v", gotMap, wantJSON)
	}

	t.Errorf("Objects structure or content does not match:\n%s", friendly)
}
