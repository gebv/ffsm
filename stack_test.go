package ffsm

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack_AddAndGet(t *testing.T) {
	s := make(Stack)
	s.Add(State("a"), State("b"), nil)
	assert.NotEmpty(t, s.Get(State("a"), State("b")))
	assert.Len(t, s.Get(State("a"), State("b")), 1)

	s.Add(State("a"), State("b"), nil)
	assert.NotEmpty(t, s.Get(State("a"), State("b")))
	assert.Len(t, s.Get(State("a"), State("b")), 2)

	s.Add(State("c"), State("d"), nil).
		Add(State("c"), State("d"), nil)
	assert.NotEmpty(t, s.Get(State("a"), State("b")))
	assert.Len(t, s.Get(State("a"), State("b")), 2)
	assert.NotEmpty(t, s.Get(State("c"), State("d")))
	assert.Len(t, s.Get(State("c"), State("d")), 2)

	assert.Empty(t, s.Get(State("not"), State("exists")))
	assert.Len(t, s.Get(State("not"), State("exists")), 0)
}

func TestAsync_Get(t *testing.T) {
	r := make(Stack)
	r.Add(State("a"), State("b"), nil)
	r.Add(State("a"), State("b"), nil)
	r.Add(State("b"), State("c"), nil)
	r.Add(State("c"), State("a"), nil)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				actions := r.Get(State("not exists"), State("not exists"))
				if len(actions) != 0 {
					t.Fatal("not expected len actions")
				}
			}()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				actions := r.Get(State("not exists"), State("not exists"))
				if len(actions) != 0 {
					t.Fatal("not expected len actions")
				}
			}()
		}
	}()
	wg.Wait()
}
