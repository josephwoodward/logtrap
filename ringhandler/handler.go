package ringhandler

import (
	"container/ring"
	"context"
	"log/slog"
	"sync"
)

type Options struct {
	// Number of per request logs buffered that will be flushed in the event of an error
	Size int

	// Attribute to index logs on
	AttrFilter string

	Level slog.Leveler
}

type RingHandler struct {
	inner slog.Handler
	opts  *Options
	buf   *ring.Ring
	mu    sync.Mutex
}

func NewRingHandler(handler slog.Handler, opts *Options) *RingHandler {
	if opts == nil {
		opts = &Options{}
	}

	return &RingHandler{
		inner: handler,
		opts:  opts,
		buf:   ring.New(opts.Size),
	}
}

// Enabled implements slog.Handler.
func (r *RingHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// WithAttrs implements slog.Handler.
func (r *RingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return r
}

// WithGroup implements slog.Handler.
func (r *RingHandler) WithGroup(name string) slog.Handler {
	return r
}

func (r *RingHandler) Handle(ctx context.Context, record slog.Record) error {
	switch record.Level {
	// case r.opts.Level:
	// r.mu.Lock()
	// defer r.mu.Unlock()
	// r.buf.Value = record.Clone()
	// r.buf = r.buf.Next()
	case slog.LevelError:
		// flush
		r.buf.Do(func(a any) {
			if s, ok := a.(slog.Record); ok {
				_ = r.inner.Handle(ctx, s)
			}
		})

		return r.inner.Handle(ctx, record)
	default:
		r.mu.Lock()
		defer r.mu.Unlock()
		r.buf.Value = record.Clone()
		r.buf = r.buf.Next()
	}

	return nil

	// return r.inner.Handle(ctx, record)
}
