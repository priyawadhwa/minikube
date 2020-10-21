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
	ctx     context.Context
	spans   map[string]trace.Span
	cleanup func()
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

// Cleanup ends the span and flushes all data
func Cleanup() {
	if t.cleanup == nil {
		return
	}
	t.cleanup()
}

func initGCPTracer() error {
	projectID := os.Getenv("MINIKUBE_TRACER_GCP_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("gcp tracer requires a valid GCP project id set via the MINIKUBE_TRACER_GCP_PROJECT_ID env variable")
	}
	// header := "X-Cloud-Trace-Context:105445aa7843bc8bf206b12000100000/1;o=1"
	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// client := &http.Client{Transport: tr}

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
	id, err := trace.IDFromHex("8fbb2742dffa3530a94ac1f32e76dca5")
	if err != nil {
		return errors.Wrap(err, "getting id from hex")
	}
	link := trace.Link{
		SpanContext: trace.SpanContext{
			TraceID: id,
		},
	}
	ctx, span := t.Tracer.Start(context.Background(), "minikube_start", trace.LinkedTo(link.SpanContext))
	t.ctx = ctx
	t.cleanup = func() {
		span.End()
		exporter.Flush()
		time.Sleep(2 * time.Second)
	}
	return nil
}

func startSpan(name string) {
	if t.Tracer == nil {
		return
	}
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
	span.End()
}
