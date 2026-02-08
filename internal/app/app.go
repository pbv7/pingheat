package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/internal/exporter"
	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/pbv7/pingheat/internal/ping"
	"github.com/pbv7/pingheat/internal/pprof"
	"github.com/pbv7/pingheat/internal/ui"
)

// runner emits ping samples until the context is cancelled.
type runner interface {
	Run(ctx context.Context, samples chan<- ping.Sample) error
}

// metricsExporter publishes metrics updates and serves them over HTTP.
type metricsExporter interface {
	Start(ctx context.Context) error
	Update(stats metrics.Stats)
}

// profiler exposes runtime profiling endpoints.
type profiler interface {
	Start(ctx context.Context) error
}

// program is the minimal Bubble Tea interface used by the app.
type program interface {
	Run() (tea.Model, error)
	Quit()
}

// programFactory builds a UI program from a model.
type programFactory func(tea.Model) program

// App orchestrates all components of pingheat.
type App struct {
	config config.Config

	// Components
	runner   runner
	engine   *metrics.Engine
	exporter metricsExporter
	pprof    profiler
	program  programFactory

	// Channels
	samples    chan ping.Sample
	uiSamples  chan ping.Sample
	metricsOut chan metrics.Stats
	errors     chan error
}

// New creates a new App instance.
func New(cfg config.Config) *App {
	app := &App{
		config:     cfg,
		runner:     ping.NewRunner(cfg.Target, cfg.Interval),
		engine:     metrics.NewEngine(),
		program:    newProgram,
		samples:    make(chan ping.Sample, 100),
		uiSamples:  make(chan ping.Sample, 100),
		metricsOut: make(chan metrics.Stats, 10),
		errors:     make(chan error, 10),
	}

	if cfg.ExporterEnabled {
		app.exporter = exporter.NewExporter(cfg.ExporterAddr, cfg.Target)
	}

	if cfg.PprofEnabled {
		app.pprof = pprof.NewServer(cfg.PprofAddr)
	}

	return app
}

// newProgram creates the default Bubble Tea program.
func newProgram(model tea.Model) program {
	return tea.NewProgram(model, tea.WithAltScreen())
}

// Run starts the application.
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if a.program == nil {
		a.program = newProgram
	}

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Start pprof server if enabled
	if a.pprof != nil {
		go func() {
			if err := a.pprof.Start(ctx); err != nil {
				a.errors <- fmt.Errorf("pprof server: %w", err)
			}
		}()
	}

	// Start exporter if enabled
	if a.exporter != nil {
		go func() {
			if err := a.exporter.Start(ctx); err != nil {
				a.errors <- fmt.Errorf("exporter: %w", err)
			}
		}()
	}

	// Start ping runner
	go func() {
		if err := a.runner.Run(ctx, a.samples); err != nil {
			a.errors <- fmt.Errorf("ping runner: %w", err)
		}
		close(a.samples)
	}()

	// Start distributor
	go a.distribute(ctx)

	// Create and run UI
	model := ui.NewModel(a.config, a.uiSamples, a.metricsOut)
	program := a.program(model)

	// Run UI in a goroutine so we can cancel it
	done := make(chan error, 1)
	go func() {
		_, err := program.Run()
		done <- err
		cancel()
	}()

	// Wait for completion
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		program.Quit()
		// Wait for UI goroutine to fully terminate with timeout
		select {
		case err := <-done:
			return err
		case <-time.After(5 * time.Second):
			return fmt.Errorf("UI failed to shut down within 5 seconds")
		}
	case err := <-a.errors:
		program.Quit()
		// Wait for UI to shut down with timeout and capture any shutdown errors
		select {
		case uiErr := <-done:
			if uiErr != nil {
				return fmt.Errorf("original error: %w; failed to shutdown UI: %v", err, uiErr)
			}
			return err
		case <-time.After(5 * time.Second):
			return fmt.Errorf("original error: %w; UI failed to shut down within 5 seconds", err)
		}
	}
}

// distribute fans out samples to consumers.
func (a *App) distribute(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(a.uiSamples)
			close(a.metricsOut)
			return
		case sample, ok := <-a.samples:
			if !ok {
				close(a.uiSamples)
				close(a.metricsOut)
				return
			}

			// Send to UI (non-blocking)
			select {
			case a.uiSamples <- sample:
			default:
				// UI buffer full, skip
			}

			// Update metrics
			a.engine.Add(sample)
			stats := a.engine.Stats()

			// Send to metrics channel (non-blocking)
			select {
			case a.metricsOut <- stats:
			default:
				// Metrics buffer full, skip
			}

			// Update exporter if enabled
			if a.exporter != nil {
				a.exporter.Update(stats)
			}
		}
	}
}
