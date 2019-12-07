# FFSM

![CI Status](https://github.com/gebv/ffsm/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebv/ffsm)](https://goreportcard.com/report/github.com/gebv/ffsm)
[![codecov](https://codecov.io/gh/gebv/ffsm/branch/master/graph/badge.svg)](https://codecov.io/gh/gebv/ffsm)

Finite state machine (FSM) written in Go. It is a low-level primitive for more complex solutions.

Working code [see more in tests](fsm_test.go)

## Features

* dispatcher is thread-safe
* implemented `prometheus.Collector`

## Examples

The following is example code.

```golang
// FSM transition service
s := &door{}

// workflow or stateflow
wf := make(Stack).
  Add(CloseDoor, OpenDoor, s.AccessOnlyBob). // bob only have access to door (sets payload via context).
  Add(OpenDoor, CloseDoor, s.Empty) // anyone can close the door.

// setup FSM with initial state
fsm := NewFSM(wf, CloseDoor)
// or sets state (thread-safe)
fsm.SetState(CloseDoor)

// send a transition request (thread-safe)
errCh, _ := fsm.Dispatch(ctx, OpenDoor)

// (optional) waiting until the processing of the current transition is completed
err := <- errCh
if err != nil {
  // handle the transition error
}

// final state (thread-safe)
fsm.State()
```
