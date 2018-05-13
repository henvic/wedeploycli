// Copyright 2016-present Liferay, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonlib

import "testing"

type AssertJSONMarshalProvider struct {
	Want string
	Got  interface{}
	Pass bool
}

type xStruct struct {
	X string `json:"x"`
}

var empty struct{}

var AssertJSONMarshalCases = []AssertJSONMarshalProvider{
	{``, nil, false},
	{`null`, nil, true},
	{`xnull`, nil, false},
	{`"x"`, "x", true},
	{`"x"`, "y", false},
	{`{}`, empty, true},
	{`["y", "z"]`, []string{"y", "z"}, true},
	{`["y", "z"]`, map[string]string{"x": "y"}, false},
	{`null`, map[string]string{"x": "y"}, false},
	{`{4: "y"}`, map[int]string{4: "y"}, false},
	{`{"x": "y"}`, xStruct{"y"}, true},
	{`["y", "z"]`, []string{"x", "y", "z"}, false},
}

func TestAssertJSONMarshal(t *testing.T) {
	for _, c := range AssertJSONMarshalCases {
		var mockTest = &testing.T{}
		AssertJSONMarshal(mockTest, c.Want, c.Got)

		if mockTest.Failed() == c.Pass {
			t.Errorf("Mock test did not meet passing status = %v assertion", c.Pass)
		}
	}
}
