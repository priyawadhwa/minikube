package register

import (
	"context"
	"fmt"
	"os"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
)

type mktracer struct {
	trace.Tracer
	ctx       context.Context
	finalSpan trace.Span
	spans     map[string]trace.Span
}

var (
	t mktracer
)

// InitializeTracer intializes the tracer
func InitializeTracer(t string) error {
	switch t {
	case "gcp":
		return initGCPTracer()
	case "":
		return nil
	}
	return fmt.Errorf("%s is not a valid tracer", t)
}

func Cleanup() {
	if t.finalSpan == nil {
		return
	}
	fmt.Println("ending final span")
	t.finalSpan.End()
	time.Sleep(2 * time.Second)
}

func initGCPTracer() error {
	projectID := os.Getenv("MINIKUBE_TRACER_GCP_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("gcp tracer requires a valid GCP project id set via the MINIKUBE_TRACER_GCP_PROJECT_ID env variable")
	}
	exporter, err := texporter.NewExporter(texporter.WithProjectID(projectID))

	if err != nil {
		return errors.Wrap(err, "getting exporter")
	}
	tp, err := sdktrace.NewProvider(sdktrace.WithSyncer(exporter))
	if err != nil {
		return errors.Wrap(err, "new provider")
	}
	tp.ApplyConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()})
	global.SetTraceProvider(tp)
	t = mktracer{
		Tracer: global.TraceProvider().Tracer("container-tools"),
		spans:  map[string]trace.Span{},
	}
	ctx, span := t.Tracer.Start(context.Background(), "minikube_start")
	t.ctx = ctx
	t.finalSpan = span
	return nil
}

func startSpan(name string) {
	if t.Tracer == nil {
		return
	}
	fmt.Println("Starting", name)
	_, span := t.Start(t.ctx, name)
	t.spans[name] = span
}

func endSpan(name string) {
	if t.Tracer == nil {
		return
	}
	span, ok := t.spans[name]
	if !ok {
		return
	}
	fmt.Println("ending", name)
	span.End()
}
