package dummy

var (
	apiPort = 0
)

// Make it possible to explicitly set the port that the fake-juju API
// server will be listening to, instead of using a random one.
func SetAPIPort(port int) {
	apiPort = port
}
