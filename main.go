package main

import (
	"log/slog"
	"os"

	. "github.com/josephwoodward/log-ring/ringhandler"
)

func main() {

	handler := slog.NewJSONHandler(os.Stdout, nil)
	opts := &Options{
		Size:       10,
		Level:      slog.LevelInfo,
		AttrFilter: "RequestId",
	}
	log := slog.New(NewRingHandler(handler, opts))
	log.Debug("Debug 1", "RequestId", "123")
	log.Debug("Debug 2", "RequestId", "456")

	log.Debug("Debug 4", "RequestId", "123")
	log.Debug("Debug 5", "RequestId", "456")

	log.Warn("Warning...", "RequestId", "123")
	log.Error("Boom!", "RequestId", "123")
}
