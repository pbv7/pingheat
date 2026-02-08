package app

import (
	"context"
	"errors"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/pbv7/pingheat/internal/ping"
)

type stubRunner struct {
	err error
}

func (r *stubRunner) Run(ctx context.Context, samples chan<- ping.Sample) error {
	if r.err != nil {
		return r.err
	}
	<-ctx.Done()
	return nil
}

type stubExporter struct {
	startErr error
	updates  int
}

func (e *stubExporter) Start(ctx context.Context) error {
	if e.startErr != nil {
		return e.startErr
	}
	<-ctx.Done()
	return nil
}

func (e *stubExporter) Update(stats metrics.Stats) {
	e.updates++
}

type stubProfiler struct {
	startErr error
}

func (p *stubProfiler) Start(ctx context.Context) error {
	if p.startErr != nil {
		return p.startErr
	}
	<-ctx.Done()
	return nil
}

type stubProgram struct {
	block      chan struct{}
	runErr     error
	quitCalled bool
	once       sync.Once
}

func (p *stubProgram) Run() (tea.Model, error) {
	<-p.block
	return nil, p.runErr
}

func (p *stubProgram) Quit() {
	p.quitCalled = true
	p.once.Do(func() {
		close(p.block)
	})
}

func newTestApp(r runner, e metricsExporter, p profiler, prog *stubProgram) *App {
	return &App{
		config:     config.DefaultConfig(),
		runner:     r,
		engine:     metrics.NewEngine(),
		exporter:   e,
		pprof:      p,
		program:    func(tea.Model) program { return prog },
		samples:    make(chan ping.Sample, 1),
		uiSamples:  make(chan ping.Sample, 1),
		metricsOut: make(chan metrics.Stats, 1),
		errors:     make(chan error, 1),
	}
}

func TestRunReturnsRunnerError(t *testing.T) {
	errRunner := errors.New("runner failed")
	prog := &stubProgram{block: make(chan struct{})}
	app := newTestApp(&stubRunner{err: errRunner}, nil, nil, prog)

	err := app.Run()

	if !errors.Is(err, errRunner) {
		t.Fatalf("expected runner error, got %v", err)
	}
	if !prog.quitCalled {
		t.Fatalf("expected program Quit to be called")
	}
}

func TestRunReturnsExporterError(t *testing.T) {
	errExporter := errors.New("exporter failed")
	prog := &stubProgram{block: make(chan struct{})}
	app := newTestApp(&stubRunner{}, &stubExporter{startErr: errExporter}, nil, prog)

	err := app.Run()

	if !errors.Is(err, errExporter) {
		t.Fatalf("expected exporter error, got %v", err)
	}
	if !prog.quitCalled {
		t.Fatalf("expected program Quit to be called")
	}
}

func TestRunReturnsPprofError(t *testing.T) {
	errProfiler := errors.New("pprof failed")
	prog := &stubProgram{block: make(chan struct{})}
	app := newTestApp(&stubRunner{}, nil, &stubProfiler{startErr: errProfiler}, prog)

	err := app.Run()

	if !errors.Is(err, errProfiler) {
		t.Fatalf("expected pprof error, got %v", err)
	}
	if !prog.quitCalled {
		t.Fatalf("expected program Quit to be called")
	}
}

func TestRunReturnsProgramError(t *testing.T) {
	errProgram := errors.New("program failed")
	prog := &stubProgram{block: make(chan struct{}), runErr: errProgram}
	close(prog.block)
	app := newTestApp(&stubRunner{}, nil, nil, prog)

	err := app.Run()
	if !errors.Is(err, errProgram) {
		t.Fatalf("expected program error, got %v", err)
	}
}
