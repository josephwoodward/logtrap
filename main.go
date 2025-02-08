package main

import (
	"log/slog"
	"os"

	. "github.com/josephwoodward/log-ring/slogring"
)

func main() {

	handler := slog.NewJSONHandler(os.Stdout, nil)
	opts := &HandlerOptions{
		TailSize:  10,
		TailLevel: slog.LevelInfo,
		AttrKey:   "RequestId",
	}
	log := slog.New(NewHandler(handler, opts))

	log.Debug("Debug 1", "RequestId", "123")
	log.Debug("Debug 2", "RequestId", "456")

	log.Debug("Debug 2", "RequestId", "123")
	log.Debug("Debug 5", "RequestId", "456")

	log.Info("Info 1", "RequestId", "123")
	log.Info("Info 2", "RequestId", "123")

	log.Warn("Warning 1", "RequestId", "123")
	log.Error("Boom!", "RequestId", "123")
}
