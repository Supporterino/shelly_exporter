package exporter

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const landingPage = `<html>
     <head><title>Shelly Exporter</title></head>
     <body>
     <h1>Shelly Exporter</h1>
     <p><a href=/metrics>Metrics</a></p>
     </body>
     </html>`

// NewHandler builds the exporter's HTTP handler.
func NewHandler(reg prometheus.Gatherer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", landingHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", healthHandler)
	return mux
}

func landingHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, landingPage); err != nil {
		slog.Debug("Failed to write landing page", slog.Any("error", err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, "Server is healthy"); err != nil {
		slog.Debug("Failed to write health response", slog.Any("error", err))
	}
}
