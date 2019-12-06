package ffsm

// AbstractMessageOfChannel backet for abstract messages of channel.
type AbstractMessageOfChannel interface{}

// Channel with abstract messages.
type Channel chan AbstractMessageOfChannel

// Close closes the channel.
func (c Channel) Close() (closed bool) {
	defer func() {
		if recover() != nil {
			closed = false
		}
	}()

	close(c)
	return true
}

// Send send a new message to the channel.
// Returns true if channel is not closed and successful send message
func (c Channel) Send(msg AbstractMessageOfChannel) (failed bool) {
	defer func() {
		if recover() != nil {
			failed = false
		}
	}()

	select {
	case c <- msg:
	default:
		return false
	}

	return true
}
