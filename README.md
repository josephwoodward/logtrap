# LogTrap

Save log quota and focus on the logs that matter with LogTrap, a handler for Go's `log/slog` package that uses a ring buffer to capture logs and only flush them upon receiving an error.

## Usage

```go
opts := &logtrap.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "request_id", FlushLevel: slog.LevelError}
handler := logtrap.NewHandler(slog.NewTextHandler(os.Stdout, nil), opts)

// act
logger := slog.New(handler)
logger.Debug("should not log")
logger.Info("should not log")
logger.Warn("should log")
```
