// Internal commands used by the FakeJujuRunner

package service

// Commands that the runner uses internally to control its execution
// flow.
type command struct {

	// Numerical identifier for the command to execute
	code int

	// Channel that will be sent "nil" or an error object once the
	// command completes (successfully or unsuccessfully). The
	// invoker of the command can use it to get notified of
	// completion.
	done chan error
}

func newCommand(code int) *command {
	return &command{
		code: code,
		done: make(chan error, 1),
	}
}

// Numerical constants identifying possible internal command codes
// that can be sent to the FakeJujuRunner.commands channel.
const (
	commandCodeStop = iota // Used when stopping the runner main loop
)
