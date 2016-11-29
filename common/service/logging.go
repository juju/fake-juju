// Logging configuration

package service

import (
	"io"
	"github.com/juju/loggo"
)

var (
	log = loggo.GetLogger("")
)

func setupLogging(output io.Writer, level loggo.Level) {

	loggo.ResetWriters()
	loggo.RegisterWriter("default", loggo.NewColorWriter(output))

	log.SetLogLevel(level)
}
