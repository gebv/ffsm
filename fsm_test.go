package ffsm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_FSM_Simple(t *testing.T) {
	wf := make(Stack).
		Add(CloseDoor, OpenDoor).
		Add(OpenDoor, CloseDoor).
		Add(OpenDoor, OpenDoor)

	fsm := NewFSM(wf, CloseDoor)
	err := fsm.Dispatch(context.Background(), OpenDoor)
	assert.NoError(t, err) // successful
	assert.Equal(t, OpenDoor, fsm.State())

	err = fsm.Dispatch(context.Background(), OpenDoor)
	assert.NoError(t, err) // successful

	err = fsm.Dispatch(context.Background(), CloseDoor)
	assert.NoError(t, err) // successful

	err = fsm.Dispatch(context.Background(), CloseDoor)
	assert.Error(t, err) // failed because not exists transition CloseDoor to CloseDoor
}

func Benchmark_FSM_Simple(b *testing.B) {
	wf := make(Stack).
		Add(CloseDoor, OpenDoor, nil).
		Add(OpenDoor, CloseDoor, nil)

	fsm := NewFSM(wf, CloseDoor)
	ctx := context.Background()
	invertState := func(in State) State {
		if in.Match(OpenDoor) {
			return CloseDoor
		}

		if in.Match(CloseDoor) {
			return OpenDoor
		}
		return UnknownState
	}
	b.ResetTimer()

	var done chan error
	var next State
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		fsm.SetState(CloseDoor)
		next = invertState(fsm.State())
		b.StartTimer()

		done, _ = fsm.AsyncDispatch(ctx, next)
		err := <-done
		if err != nil {
			b.Fatal("dispatch with err", err)
		}

		b.StopTimer()
		if fsm.State() != next {
			b.Fatalf("want %q, got %q", next, fsm.State())
		}
		b.StartTimer()
	}
	b.ReportAllocs()
}

func Test_FSM_TransitionWithHandlers(t *testing.T) {
	door := &door{}
	var NotExistsState = State("not exists")

	tests := []struct {
		name              string
		wf                Stack
		ctx               context.Context
		initState         State
		pushState         State
		finiteState       State
		err               error
		cancelCtx         bool
		doNotWaitComplete bool // do not wait for dispatcher to complete
	}{
		{
			name:        "forbiddenForAnonymous_emptyContext",
			wf:          make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.Background(),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			err:         errors.New("access denied"), // receive from WF
		},
		{
			name:        "accessOnlyBob_viaContext_successful",
			wf:          make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.WithValue(context.Background(), "__name", "bob"),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: OpenDoor, // successful
			err:         nil,
		},
		{
			name:        "accessOnlyBob_viaContext_fail",
			wf:          make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.WithValue(context.Background(), "__name", "alise"),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			err:         errors.New("access denied"), // forbidden for alise, only for bob
		},
		{
			name:        "contextCanceled_afterDispatch",
			wf:          make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.Background(),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			cancelCtx:   true,
			err:         errors.New("context canceled"),
		},
		{
			name:              "notWaitForTheResult",
			wf:                make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:               context.Background(),
			initState:         CloseDoor,
			pushState:         OpenDoor,
			finiteState:       CloseDoor,
			err:               nil, // did not receive
			doNotWaitComplete: true,
		},
		{
			name:              "notWaitForTheResultAndContextCanceled",
			wf:                make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:               context.Background(),
			initState:         CloseDoor,
			pushState:         OpenDoor,
			finiteState:       CloseDoor,
			cancelCtx:         true,
			err:               nil, // did not receive
			doNotWaitComplete: true,
		},
		{
			name:              "payloadInTheContext",
			wf:                make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:               context.WithValue(context.Background(), "__name", "bob"),
			initState:         CloseDoor,
			pushState:         OpenDoor,
			finiteState:       CloseDoor, // did not receive
			err:               nil,
			doNotWaitComplete: true,
		},

		{
			name: "pipelineOfContext",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob). // set the value in WF to context
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.Background(), // push with empty context
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: OpenDoor,
			err:         nil,
		},

		{
			name: "abortedOfOpeningDoor",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob).
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob).
				Add(CloseDoor, OpenDoor, door.AbortOpen), // aborted of opening door
			ctx:         context.Background(),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			err:         errors.New("abort open door"),
		},

		{
			name: "notExistsInitState",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob).
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.Background(),
			initState:   NotExistsState,
			pushState:   OpenDoor,
			finiteState: NotExistsState,      // becaouse init state NotExistsState
			err:         ErrNotRegTransition, // because not exists transition NotExistsState->OpenDoor
		},

		{
			name: "notExistsPushState",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob).
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx:         context.Background(),
			initState:   CloseDoor,
			pushState:   NotExistsState,
			finiteState: CloseDoor,
			err:         ErrNotRegTransition, // because not exists transition CloseDoor->NotExistsState
		},

		{
			name: "panicInTheSomeHandlerOfTransition",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob).
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob).
				Add(CloseDoor, OpenDoor, door.Panic),
			ctx:         context.Background(),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			err:         errors.New(`dispatcher panic: door.Panic ("close"=>"open" #2)`),
		},

		{
			name: "contextDeadline",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob).
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 0)
				return ctx
			}(),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			err:         errors.New("context deadline exceeded"),
		},
		{
			name: "contextCanceled_beforeDispatch",
			wf: make(Stack).
				Add(CloseDoor, OpenDoor, door.IfAnonymThenBob).
				Add(CloseDoor, OpenDoor, door.AccessOnlyBob),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			initState:   CloseDoor,
			pushState:   OpenDoor,
			finiteState: CloseDoor,
			err:         errors.New("context canceled"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsm := NewFSM(tt.wf, tt.initState)

			errCh, cancel := fsm.AsyncDispatch(tt.ctx, tt.pushState)
			if tt.cancelCtx {
				cancel()
			}

			var err error
			if !tt.doNotWaitComplete {
				err = <-errCh
			}

			if tt.err != nil {
				assert.Error(t, err)

				// TODO: it is correctly?
				assert.Contains(t, err.Error(), tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.finiteState, fsm.State(), "inite state")
		})
	}

	time.Sleep(time.Second)
}

func Test_FSM_FullState_ConcurrentDispatch(t *testing.T) {
	door := &door{}
	wf := make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob).
		Add(OpenDoor, CloseDoor, door.Empty)
	ctx := context.WithValue(context.Background(), "__name", "bob")
	fsm := NewFSM(wf, CloseDoor)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				fsm.Size()
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 100; n++ {
				done, _ := fsm.AsyncDispatch(ctx, OpenDoor)
				<-done
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 100; n++ {
				fsm.Dispatch(ctx, CloseDoor)
			}
		}()
	}

	assert.NotEqual(t, 0, fsm.Size())
	wg.Wait()
	assert.EqualValues(t, 0, fsm.Size())
}

func Benchmark_FSM_TransitionWithHandlers(b *testing.B) {
	door := &door{}
	wf := make(Stack).
		Add(CloseDoor, OpenDoor, door.AccessOnlyBob).
		Add(OpenDoor, CloseDoor, door.Empty)
	ctx := context.WithValue(context.Background(), "__name", "bob")
	invertState := func(in State) State {
		if in.Match(OpenDoor) {
			return CloseDoor
		}

		if in.Match(CloseDoor) {
			return OpenDoor
		}
		return UnknownState
	}
	fsm := NewFSM(wf, CloseDoor)
	b.ResetTimer()
	var done chan error
	var next State
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		fsm.SetState(CloseDoor)
		next = invertState(fsm.State())
		b.StartTimer()

		done, _ = fsm.AsyncDispatch(ctx, next)
		err := <-done
		if err != nil {
			b.Fatal("dispatch with err", err)
		}

		b.StopTimer()
		if fsm.State() != next {
			b.Fatalf("want %q, got %q", next, fsm.State())
		}
		b.StartTimer()
	}
	b.ReportAllocs()
}

const (
	OpenDoor   State = "open"
	TokTokDoor State = "toktok"
	CloseDoor  State = "close"
)

type door struct {
}

func (d door) IfAnonymThenBob(ctx context.Context) (context.Context, error) {
	if ctx.Value("__name") != nil {
		return ctx, nil
	}
	return context.WithValue(ctx, "__name", "bob"), nil
}

func (d door) Panic(ctx context.Context) (context.Context, error) {
	panic("door.Panic")
	return ctx, nil
}

func (d door) Empty(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (d door) AccessOnlyBob(ctx context.Context) (context.Context, error) {
	name, ok := ctx.Value("__name").(string)
	if !ok {
		return ctx, errors.New("access denied")
	}
	if name != "bob" {
		return ctx, errors.New("access denied")
	}
	return ctx, nil
}

func (d door) AbortOpen(ctx context.Context) (context.Context, error) {
	return ctx, errors.New("abort open door")
}
