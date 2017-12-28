/*
Copyright (c) 2016-present, Liferay Inc. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of Liferay, Inc. nor the names of its contributors may
be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

package urilib

import "testing"

var dataProvider = []struct {
	paths    []string
	expected string
}{
	{
		[]string{"", "foo"},
		"foo",
	},
	{
		[]string{"", "foo/"},
		"foo",
	},
	{
		[]string{"foo", ""},
		"foo",
	},
	{
		[]string{"foo/", ""},
		"foo",
	},
	{
		[]string{"foo", "/bar"},
		"foo/bar",
	},
	{
		[]string{"foo", "bar"},
		"foo/bar",
	},
	{
		[]string{"foo/", "/bar"},
		"foo/bar",
	},
	{
		[]string{"foo/", "bar"},
		"foo/bar",
	},
	{
		[]string{"foo", "/bar", "bazz"},
		"foo/bar/bazz",
	},
	{
		[]string{"http://localhost:123", ""},
		"http://localhost:123",
	},
	{
		[]string{"http://localhost:123", "/foo", "bah"},
		"http://localhost:123/foo/bah",
	},
}

func TestResolvePath(t *testing.T) {
	for _, test := range dataProvider {
		got := ResolvePath(test.paths...)
		if got != test.expected {
			t.Errorf("For %q got %q; expected %q", test.paths, got, test.expected)
		}
	}
}
