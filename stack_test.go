package ffsm

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack_AddAndGet(t *testing.T) {
	s := make(Stack)
	s.Add("a", "b", nil)
	assert.NotEmpty(t, s.Get("a", "b"))
	assert.Len(t, s.Get("a", "b"), 1)

	s.Add("a", "b", nil)
	assert.NotEmpty(t, s.Get("a", "b"))
	assert.Len(t, s.Get("a", "b"), 2)

	s.Add("c", "d", nil).
		Add("c", "d", nil)
	assert.NotEmpty(t, s.Get("a", "b"))
	assert.Len(t, s.Get("a", "b"), 2)
	assert.NotEmpty(t, s.Get("c", "d"))
	assert.Len(t, s.Get("c", "d"), 2)

	assert.Empty(t, s.Get("not", "exsts"))
	assert.Len(t, s.Get("not", "exsts"), 0)
}

func TestAsync_Get(t *testing.T) {
	r := make(Stack)
	r.Add("a", "b", nil)
	r.Add("a", "b", nil)
	r.Add("b", "c", nil)
	r.Add("c", "a", nil)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				actions := r.Get("no exists", "not exists")
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
				actions := r.Get("no exists", "not exists")
				if len(actions) != 0 {
					t.Fatal("not expected len actions")
				}
			}()
		}
	}()
	wg.Wait()
}
