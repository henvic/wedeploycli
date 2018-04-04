package ack

import (
	"math"
	"sync"
)

// Waiter for registering acks to be fulfilled.
type Waiter struct {
	counter     int
	counterLock sync.Mutex

	message map[int](chan string)
	lock    sync.RWMutex
}

// Next gets a new ID for an ack message.
func (w *Waiter) Next() int {
	w.counterLock.Lock()

	if w.counter == math.MaxInt32 {
		w.counter = -1
	}

	w.counter++
	c := w.counter

	w.counterLock.Unlock()
	return c
}

// Set message.
func (w *Waiter) Set(id int, msg chan string) {
	w.lock.Lock()

	if w.message == nil {
		w.message = map[int](chan string){}
	}

	w.message[id] = msg
	w.lock.Unlock()
}

// Delete message.
func (w *Waiter) Delete(id int) {
	w.lock.Lock()
	delete(w.message, id)
	w.lock.Unlock()
}

// Load a stored ack, or nil if no value is present. The ok result indicates whether a value was found.
func (w *Waiter) Load(id int) (chan string, bool) {
	w.lock.RLock()
	waiter, ok := w.message[id]
	w.lock.RUnlock()
	return waiter, ok
}

// Size of the map.
func (w *Waiter) Size() int {
	w.lock.RLock()
	s := len(w.message)
	w.lock.RUnlock()
	return s
}
