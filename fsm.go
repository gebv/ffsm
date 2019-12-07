package ffsm

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	toDispatchCap = 128
)

// Dispatcher dispatcher of finite state machine.
type Dispatcher func(ctx context.Context, next State) (chan error, context.CancelFunc)

// NewFSM returns new finite state machine with initial state.
func NewFSM(wf Stack, initState State) *FSM {
	e := &FSM{
		wf:         wf,
		toDispatch: make(chan *messageToDispatch, toDispatchCap),
		state:      initState,
		mActionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:      "ffsm_action_duration_ms",
				Help:      "Duration of a single action.",
				Subsystem: "ffsm",
				Buckets: []float64{
					float64(time.Millisecond * 100),
					float64(time.Millisecond * 300),
					float64(time.Millisecond * 700),
					float64(time.Millisecond * 1000),
				},
			},
			[]string{"ffsm"},
		),
		mTotalDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:      "ffsm_total_duration_ms",
				Help:      "Duration of a total dispatch.",
				Subsystem: "ffsm",
				Buckets: []float64{
					float64(time.Millisecond * 500),
					float64(time.Millisecond * 1000),
					float64(time.Millisecond * 2000),
					float64(time.Millisecond * 5000),
				},
			},
			[]string{"ffsm"},
		),
		mActionRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "ffsm_exec_action_total",
				Help:      "Number of execute to actions.",
				Subsystem: "ffsm",
			},
			[]string{"ffsm"},
		),
		mTotalRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "ffsm_dispatch_total",
				Help:      "Number of dispatch.",
				Subsystem: "ffsm",
			},
			[]string{"ffsm"},
		),
	}
	e.wg.Add(1)
	go e.runDispatcher()
	return e
}

// FSM finite state machine.
type FSM struct {
	wf         Stack
	state      State
	stateMutex sync.RWMutex
	wg         sync.WaitGroup
	toDispatch chan *messageToDispatch

	name string

	mActionDuration *prometheus.HistogramVec
	mTotalDuration  *prometheus.HistogramVec
	mActionRequest  *prometheus.CounterVec
	mTotalRequest   *prometheus.CounterVec
}

// State returns current state.
func (e *FSM) State() State {
	e.stateMutex.RLock()
	defer e.stateMutex.RUnlock()
	return e.state
}

// SetName sets name of FSM (for prometheus labels).
func (e *FSM) SetName(name string) {
	e.name = name
}

// SetState sets new state.
func (e *FSM) SetState(newState State) {
	e.stateMutex.Lock()
	e.state = newState
	e.stateMutex.Unlock()
}

type dispatcherError struct {
	Err         error
	Recover     interface{}
	DebugStack  string
	SrcState    State
	DstState    State
	IndexAction int
}

func (e dispatcherError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Recover == nil {
		return ""
	}
	return fmt.Sprintf("dispatcher panic: %v (%q=>%q #%d)", e.Recover, e.SrcState, e.DstState, e.IndexAction)
}

type resultOfActionTransition struct {
	ctx context.Context
	err error
}

func (e *FSM) runDispatcher() {
	defer e.wg.Done()

	var current State
	var actions []Procedure
	var err error
	var dispatchStart, actionStart time.Time
	for m := range e.toDispatch {
		dispatchStart = time.Now()
		err = nil
		current = e.State()
		if current.Match(UnknownState) {
			if m.done != nil {
				m.done <- ErrNotInitalState
				continue
			}
		}
		if current.Match(m.next) {
			m.done <- ErrTransitionToItSelf
			continue
		}
		actions = e.wf.Get(current, m.next)
		if actions == nil || len(actions) == 0 {
			m.done <- ErrNotRegTransition
			continue
		}

		if m.ctx.Err() != nil {
			m.done <- m.ctx.Err()
			continue
		}

		nextCtx, cancel := context.WithCancel(hydrateContextForAction(m.ctx, e.Dispatch, current, m.next))
		defer cancel()

		for _i, action := range actions {
			actionStart = time.Now()
			var actionRes = make(chan *resultOfActionTransition, 1)
			go func(ctx context.Context) {
				defer func() {
					if r := recover(); r != nil {
						actionRes <- &resultOfActionTransition{
							err: dispatcherError{
								Recover:     r,
								SrcState:    current,
								DstState:    m.next,
								IndexAction: _i,
								DebugStack:  string(debug.Stack()),
							},
						}
						return
					}
				}()

				ctx, err := action(nextCtx)
				actionRes <- &resultOfActionTransition{
					err: err,
					ctx: ctx,
				}
			}(nextCtx)

			// waiting done action
			select {
			case done := <-actionRes:
				nextCtx = done.ctx
				err = done.err
			}

			actName := fmt.Sprintf("%q -> %q #%d", current, m.next, _i)
			e.mActionDuration.WithLabelValues(actName).Observe(float64(time.Since(actionStart).Nanoseconds() / int64(time.Millisecond)))
			e.mActionRequest.WithLabelValues(actName).Inc()

			if err != nil {
				// exit transition, because there was an error on one
				// of the handlers of transition
				break
			}

			// forend actions
		}

		if err == nil {
			e.SetState(m.next)
		}
		m.done <- err

		e.mActionDuration.WithLabelValues(e.name).Observe(float64(time.Since(dispatchStart).Nanoseconds() / int64(time.Millisecond)))
		e.mTotalRequest.WithLabelValues(e.name).Inc()
	} // forend dispatch
}

// Dispatch dispatcher of finite state machine.
// Returns the channel for feedback and the function of cancel of transition context.
func (e *FSM) Dispatch(ctx context.Context, next State) (chan error, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	msg := &messageToDispatch{
		ctx:  ctx,
		next: next,
		done: make(chan error, 1),
	}
	e.toDispatch <- msg
	return msg.done, cancel
}

// Stop stops finite state machine.
func (e *FSM) Stop() {
	close(e.toDispatch)
	e.wg.Wait()
}

type messageToDispatch struct {
	ctx  context.Context
	next State
	done chan error
}
