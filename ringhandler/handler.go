package ringhandler

import (
	"container/ring"
	"context"
	"log/slog"
	"sync"
)

type HandlerOptions struct {
	// Number of per request logs buffered that will be flushed in the event of an error
	TailSize int

	// Attribute to index logs on
	AttrKey string

	Level slog.Leveler
}
type commonHandler struct {
	inner  slog.Handler
	opts   HandlerOptions
	mu     *sync.Mutex
	buffer map[string]*ring.Ring
}

type LogTailHandler struct {
	*commonHandler
}

// NewLogTailHandler creates a [LogTailHandler] that writes to the handler.
// If opts is nil, the default options are used.
func NewLogTailHandler(handler slog.Handler, opts *HandlerOptions) *LogTailHandler {
	if opts == nil {
		opts = &HandlerOptions{}
	}

	return &LogTailHandler{
		&commonHandler{
			inner:  handler,
			buffer: make(map[string]*ring.Ring),
			opts:   *opts,
			mu:     &sync.Mutex{},
		},
	}
}

// Enabled implements slog.Handler.
func (r *LogTailHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// WithAttrs implements slog.Handler.
func (r *LogTailHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return r
}

// WithGroup implements slog.Handler.
func (r *LogTailHandler) WithGroup(name string) slog.Handler {
	return r
}

func (r *LogTailHandler) Handle(ctx context.Context, record slog.Record) error {
	var key string
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == r.opts.AttrKey {
			key = a.Value.String()
			return false
		}
		return true
	})

	switch record.Level {

	// flush buffer on error
	case slog.LevelError:
		// flush
		if buf, ok := r.buffer[key]; ok {
			buf.Do(func(a any) {
				if s, ok := a.(slog.Record); ok {
					_ = r.inner.Handle(ctx, s)
				}
			})
		}

		return r.inner.Handle(ctx, record)
	// append buffer on everything else
	default:
		r.mu.Lock()
		defer r.mu.Unlock()

		// var (
		// 	buf *ring.Ring
		// 	ok  bool
		// )
		if buf, ok := r.buffer[key]; ok {
			buf.Value = record.Clone()
			buf = buf.Next()
			r.buffer[key] = buf
		} else {
			buf = ring.New(r.opts.TailSize)
			buf.Value = record.Clone()
			buf = buf.Next()
			r.buffer[key] = buf
		}
	}

	return nil

	// return r.inner.Handle(ctx, record)
}
