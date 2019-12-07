package ffsm

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_FSM_Simple1(t *testing.T) {
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

			errCh, cancel := fsm.Dispatch(tt.ctx, tt.pushState)
			if tt.cancelCtx {
				cancel()
			}

			var err error
			if !tt.doNotWaitComplete {
				err = <-errCh
			}

			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.finiteState, fsm.State(), "inite state")
		})
	}

	time.Sleep(time.Second)
}

func Benchmark_FSM_Simple1(b *testing.B) {
	door := &door{}
	wf := make(Stack).Add(CloseDoor, OpenDoor, door.AccessOnlyBob)
	ctx := context.WithValue(context.Background(), "__name", "bob")
	fsm := NewFSM(wf, CloseDoor)
	b.ResetTimer()
	var done chan error
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		fsm.SetState(CloseDoor)
		b.StartTimer()

		done, _ = fsm.Dispatch(ctx, OpenDoor)
		err := <-done
		if err != nil {
			b.Fatal("dispatch with err", err)
		}

		b.StopTimer()
		if fsm.State() != OpenDoor {
			b.Fatal("want OpenDoor")
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
