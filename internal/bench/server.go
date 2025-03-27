package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/josephwoodward/logtrap"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	inner := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	h := logtrap.NewHandler(inner, &logtrap.HandlerOptions{
		TailLevel: slog.LevelDebug,
		AttrKey:   "request_id",
	})

	logger := slog.New(h)
	logger.Info("server listening on port 8090")

	f := func(w http.ResponseWriter, r *http.Request) {
		// id := r.URL.Query().Get("request-id")
		id := time.Now().Unix()
		logger.Debug("received request", "request_id", id)
		w.Write([]byte("OK"))
		logger.Debug("writing response", "request_id", id)
	}

	f2 := func(w http.ResponseWriter, r *http.Request) {
		// id := r.URL.Query().Get("request-id")
		id := time.Now().Unix()
		logger.Debug("received request", "request_id", id)
		w.Write([]byte("OK"))
		logger.Debug("writing response", "request_id", id)
		logger.Error("failed to write response", "request_id", id)
	}

	mux := http.NewServeMux()
	mux.Handle("/log-info", http.HandlerFunc(f))
	mux.Handle("/log-error", http.HandlerFunc(f2))
	// mux.Handle("/metrics", promhttp.Handler())

	logger.Info("Listening on :8090...")
	err := http.ListenAndServe(":8090", mux)
	log.Fatal(err)
}
