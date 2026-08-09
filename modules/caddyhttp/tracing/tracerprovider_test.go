package tracing

import (
	"errors"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func Test_tracersProvider_getTracerProvider(t *testing.T) {
	tp := tracerProvider{}

	var buildCount int
	buildOpts := func() ([]sdktrace.TracerProviderOption, error) {
		buildCount++
		return nil, nil
	}

	p1, err := tp.getTracerProvider(buildOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p2, err := tp.getTracerProvider(func() ([]sdktrace.TracerProviderOption, error) {
		t.Fatal("buildOpts should not run on reuse")
		return nil, errors.New("should not run")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tp.tracerProvider == nil {
		t.Errorf("There should be tracer provider")
	}

	if p1 != p2 || p1 != tp.tracerProvider {
		t.Errorf("Expected the same tracer provider to be reused")
	}

	if tp.tracerProvidersCounter != 2 {
		t.Errorf("Tracer providers counter should equal to 2")
	}

	if buildCount != 1 {
		t.Errorf("buildOpts should run only once on create, got %d", buildCount)
	}
}

func Test_tracersProvider_getTracerProvider_buildOptsError(t *testing.T) {
	tp := tracerProvider{}

	_, err := tp.getTracerProvider(func() ([]sdktrace.TracerProviderOption, error) {
		return nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error from buildOpts")
	}

	if tp.tracerProvider != nil {
		t.Errorf("provider should remain nil after construction error")
	}
	if tp.tracerProvidersCounter != 0 {
		t.Errorf("counter should remain 0 after construction error, got %d", tp.tracerProvidersCounter)
	}
}

func Test_tracersProvider_cleanupTracerProvider(t *testing.T) {
	tp := tracerProvider{}

	noOpts := func() ([]sdktrace.TracerProviderOption, error) { return nil, nil }

	if _, err := tp.getTracerProvider(noOpts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := tp.getTracerProvider(noOpts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := tp.cleanupTracerProvider(zap.NewNop())
	if err != nil {
		t.Errorf("There should be no error: %v", err)
	}

	if tp.tracerProvider == nil {
		t.Errorf("There should be tracer provider")
	}

	if tp.tracerProvidersCounter != 1 {
		t.Errorf("Tracer providers counter should equal to 1")
	}
}
