// Copyright 2016-present Liferay, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qrange

import (
	"testing"

	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
)

func TestFromTo(t *testing.T) {
	var want = `{
		"from": 10,
		"to": 20
}`

	var got = Between(10, 20)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFromOnly(t *testing.T) {
	var want = `{"from":10}`
	var got = From(10)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestToOnly(t *testing.T) {
	var want = `{"to":20}`
	var got = To(20)
	jsonlib.AssertJSONMarshal(t, want, got)
}
