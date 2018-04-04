package ack

import (
	"math"
	"sync"
	"testing"
)

func TestWaiter(t *testing.T) {
	var w = Waiter{}

	if s := w.Size(); s != 0 {
		t.Errorf("Expected size to be 0, got %v instead", s)
	}

	next := w.Next()

	if next != 1 {
		t.Errorf("Expected next ID to be 1, got %v instead", next)
	}

	var m chan string

	w.Set(1, m)

	next = w.Next()

	if next != 2 {
		t.Errorf("Expected next ID to be 2, got %v instead", next)
	}

	switch got, ok := w.Load(1); {
	case !ok:
		t.Error("Expected message chan to be retrieved")
	case got != m:
		t.Errorf("Expected message chan doesn't matter")
	}

	w.Delete(1)

	if s := w.Size(); s != 0 {
		t.Errorf("Expected size to be 0, got %v instead", s)
	}

	if next := w.Next(); next != 3 {
		t.Errorf("Expected next ID to be 3, got %v instead", next)
	}
}

func TestWaiterConcurrency(t *testing.T) {
	var w = Waiter{}
	var m = make(chan string, 100)
	var q sync.WaitGroup
	q.Add(1)

	go func() {
		w.Set(w.Next(), m)
		m <- "hello"
		q.Done()
	}()

	q.Wait()

	if s := w.Size(); s != 1 {
		t.Errorf("Expected size to be 1, got %v instead", s)
	}

	switch gotc, ok := w.Load(1); {
	case !ok:
		t.Error("Expected message chan to be retrieved")
	default:
		if r := <-gotc; r != "hello" {
			t.Errorf(`Expected retrieved message to be "hello", got %v instead`, r)
		}
	}
}

func TestWaiterLimit(t *testing.T) {
	var w = Waiter{}

	w.counter = math.MaxInt32 - 2

	w.Next()
	w.Next()
	w.Next()
	w.Delete(math.MaxInt32)
	w.Next()

	if next := w.Next(); next != 2 {
		t.Errorf("Expected next position to be 2, got %v instead", next)
	}
}
