package ffsm

import "testing"

func TestState_Match(t *testing.T) {

	tests := []struct {
		name string
		in   State
		args State
		want bool
	}{
		{name: "empty", want: true},
		{in: State("foo"), args: State("foo"), want: true},
		{in: State("foo"), args: State("bar"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Match(tt.args); got != tt.want {
				t.Errorf("State.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
