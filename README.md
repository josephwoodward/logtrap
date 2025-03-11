# LogTrap

Save log quota and focus on the logs that matter with LogTrap, a handler for Go's `log/slog` package that uses a ring buffer to capture logs and only flush them upon receiving a log of the configured "flush level".

### Examples

Write `Error` and `Warning` logs, but only flush `Info` and `Debug` logs when an `Error` occurs:

```go
inner := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
h := logtrap.NewHandler(inner, &logtrap.HandlerOptions{
	TailSize:   10,
	TailLevel:  slog.LevelInfo,
	FlushLevel: slog.LevelError,
})
logger := slog.New(h)

logger.Debug("Not logged until error encountered")
logger.Info("Not logged until error encountered")
logger.Warn("Will log")
logger.Error("Will log and flush Info and Debug logs")

```
