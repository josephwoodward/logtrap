# Ring Log


## Usage

```go
opts := &logring.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "request_id", FlushLevel: slog.LevelError}
handler := logring.NewHandler(slog.NewTextHandler(os.Stdout, nil), opts)

// act
logger := slog.New(handler)
logger.Debug("should not log")
logger.Info("should not log")
logger.Warn("should log")
```
