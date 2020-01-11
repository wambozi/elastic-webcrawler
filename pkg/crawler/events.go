package crawler

import (
	"github.com/kataras/go-events"
	"github.com/sirupsen/logrus"
)

// Events returns an Event Emitter with listeners attached
func Events(logger *logrus.Logger) events.EventEmmiter {
	e := events.New()
	logger.Info("Creating Event Emitter.")

	e.On("READY", func(payload EventPayload) {

	})

	e.On("STARTED", func(payload ...interface{}) {
		logger.Info(payload)
	})

	e.On("ERROR", func(payload ...interface{}) {
		logger.Info(payload)
	})

	return e
}
