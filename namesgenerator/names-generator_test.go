// Licensed under Apache license 2.0
// Copyright 2013-2016 Docker, Inc.

// NOTICE: export from moby/pkg/namesgenerator/names-generator_test.go
// https://github.com/moby/moby/blob/66cfe61f71252f528ddb458d554cd241e996d9f1/pkg/namesgenerator/names-generator_test.go
// License: https://github.com/moby/moby/blob/66cfe61f71252f528ddb458d554cd241e996d9f1/LICENSE

package namesgenerator

import (
	"strings"
	"testing"
)

func TestNameFormat(t *testing.T) {
	name := GetRandomName(0)
	if !strings.Contains(name, "_") {
		t.Fatalf("Generated name does not contain an underscore")
	}
	if strings.ContainsAny(name, "0123456789") {
		t.Fatalf("Generated name contains numbers!")
	}
}

func TestNameRetries(t *testing.T) {
	name := GetRandomName(1)
	if !strings.Contains(name, "_") {
		t.Fatalf("Generated name does not contain an underscore")
	}
	if !strings.ContainsAny(name, "0123456789") {
		t.Fatalf("Generated name doesn't contain a number")
	}

}
