package tuiexporter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/httpserver"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type tuiExporter struct {
	app        *tui.TUIApp
	httpServer *http.Server
	httpPort   int
	serverOnly bool
}

func newTuiExporter(config *Config) (*tuiExporter, error) {
	var initialInterval time.Duration
	if config.FromJSONFile {
		// FIXME: When reading telemetry from a JSON file on startup, the UI will break
		//        if it runs at the same time as the UI drawing. As a workaround, wait for a second.
		initialInterval = 1 * time.Second
	}

	// Create store
	store := telemetry.NewStore(clockwork.NewRealClock())

	exporter := &tuiExporter{
		httpPort:   config.HTTPPort,
		serverOnly: config.ServerOnly,
	}

	// Only create TUI app if not in server-only mode
	if !config.ServerOnly {
		app, err := tui.NewTUIApp(store, initialInterval, config.DebugLogFilePath)
		if err != nil {
			return nil, err
		}
		exporter.app = app
	} else {
		// In server-only mode, create a minimal wrapper that just holds the store
		exporter.app = &tui.TUIApp{}
		// Inject the store directly (we'll need to add a method for this)
		// For now, we'll create the app anyway but won't run it
		app, err := tui.NewTUIApp(store, initialInterval, config.DebugLogFilePath)
		if err != nil {
			return nil, err
		}
		exporter.app = app
		fmt.Println("Running in server-only mode (TUI disabled)")
	}

	// Setup HTTP server if port is configured
	if config.HTTPPort > 0 {
		httpHandler := httpserver.NewServer(exporter.app.Store())
		exporter.httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", config.HTTPPort),
			Handler: httpHandler,
		}
	}

	return exporter, nil
}

func (e *tuiExporter) pushTraces(_ context.Context, traces ptrace.Traces) error {
	e.app.Store().AddSpan(&traces)

	return nil
}

func (e *tuiExporter) pushMetrics(_ context.Context, metrics pmetric.Metrics) error {
	e.app.Store().AddMetric(&metrics)

	return nil
}

func (e *tuiExporter) pushLogs(_ context.Context, logs plog.Logs) error {
	e.app.Store().AddLog(&logs)

	return nil
}

// Start runs the TUI exporter
func (e *tuiExporter) Start(ctx context.Context, _ component.Host) error {
	// Start TUI app only if not in server-only mode
	if !e.serverOnly {
		go func() {
			err := e.app.Run()
			if err != nil {
				fmt.Printf("error running tui app: %s\n", err)
			}
		}()
	}

	// Start HTTP server if configured
	if e.httpServer != nil {
		go func() {
			fmt.Printf("Starting HTTP API server on port %d\n", e.httpPort)
			if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("error running http server: %s\n", err)
			}
		}()
	}

	// In server-only mode, keep running (block) so the collector doesn't exit
	if e.serverOnly {
		fmt.Println("Server-only mode active. Press Ctrl+C to stop.")
		<-ctx.Done()
	}

	return nil
}

// Shutdown stops the TUI exporter
func (e *tuiExporter) Shutdown(ctx context.Context) error {
	// Stop HTTP server if running
	if e.httpServer != nil {
		if err := e.httpServer.Shutdown(ctx); err != nil {
			fmt.Printf("error shutting down http server: %s\n", err)
		}
	}

	// Stop TUI app only if not in server-only mode
	if !e.serverOnly && e.app != nil {
		return e.app.Stop()
	}

	return nil
}
