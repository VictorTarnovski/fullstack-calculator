// Command api serves expression evaluation over HTTP, configured entirely from
// CALCULATOR_-prefixed environment variables.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"calculator/internal/expression"
	"calculator/internal/httpx"
)

const (
	addrEnv           = "CALCULATOR_ADDR"
	allowedOriginsEnv = "CALCULATOR_ALLOWED_ORIGINS"
	defaultAddr       = ":8080"
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

// parseOrigins splits and validates the browser origins allowed to call the
// API. An origin must be scheme://host[:port] with no path, because that is
// exactly what a browser puts in the Origin header and CORS compares the two
// literally: a configured trailing slash or path would reject every request.
func parseOrigins(raw string) ([]string, error) {
	var origins []string

	for part := range strings.SplitSeq(raw, ",") {
		origin := strings.TrimRight(strings.TrimSpace(part), "/")
		if origin == "" {
			continue
		}

		u, err := url.Parse(origin)
		if err != nil {
			return nil, fmt.Errorf("%s: %q: %w", allowedOriginsEnv, origin, err)
		}

		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, fmt.Errorf("%s: %q must start with http:// or https://", allowedOriginsEnv, origin)
		}

		if u.Host == "" || u.Path != "" {
			return nil, fmt.Errorf("%s: %q must be scheme://host[:port] with no path", allowedOriginsEnv, origin)
		}

		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return nil, fmt.Errorf("%s is required, list the browser origins allowed to call this API", allowedOriginsEnv)
	}

	return origins, nil
}

func routes(kit *httpx.Kit) *http.ServeMux {
	mux := http.NewServeMux()

	(&expression.Handler{}).RegisterRoutes(mux, kit)

	return mux
}

// run serves POST /evaluations until SIGINT or SIGTERM, then drains in-flight
// requests. It reads CALCULATOR_ADDR (default :8080) and the required
// CALCULATOR_ALLOWED_ORIGINS from the environment.
func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	origins, err := parseOrigins(os.Getenv(allowedOriginsEnv))
	if err != nil {
		logger.Error("invalid configuration", "err", err)
		return err
	}

	kit := &httpx.Kit{
		Log:      logger,
		Classify: httpx.NewErrorMapper(expression.ClassifyError),
	}

	addr := os.Getenv(addrEnv)
	if addr == "" {
		addr = defaultAddr
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpx.CORS(origins)(routes(kit)),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	serveErr := make(chan error, 1)

	go func() {
		logger.Info("calculator api listening", "addr", addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		logger.Error("server", "err", err)
		return err
	case <-stop:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
