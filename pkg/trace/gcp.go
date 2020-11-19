/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trace

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel/api/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"k8s.io/klog"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
)

const (
	projectEnvVar    = "MINIKUBE_GCP_PROJECT_ID"
	parentSpanName   = "minikube start"
	customMetricName = "custom.googleapis.com/minikube/start_time"
)

type gcpTracer struct {
	projectID string
	traceExporter
	metricExporter
}

type traceExporter struct {
	trace.Tracer
	parentCtx context.Context
	spans     map[string]trace.Span
	cleanup   func()
}

type metricExporter struct {
	cleanup func()
}

func (t *gcpTracer) StartSpan(name string) {
	_, span := t.Tracer.Start(t.parentCtx, name)
	t.spans[name] = span
}
func (t *gcpTracer) EndSpan(name string) {
	span, ok := t.spans[name]
	if !ok {
		klog.Warningf("cannot end span %s as it was never started", name)
		return
	}
	span.End()
}

func (t *gcpTracer) Cleanup() {
	t.traceExporter.cleanup()
	t.metricExporter.cleanup()
}

func initGCPTracer() (*gcpTracer, error) {
	projectID := os.Getenv(projectEnvVar)
	if projectID == "" {
		return nil, fmt.Errorf("GCP tracer requires a valid GCP project id set via the %s env variable", projectEnvVar)
	}

	tex, err := getTraceExporter(projectID)
	if err != nil {
		return nil, errors.Wrap(err, "getting trace exporter")
	}

	return &gcpTracer{
		projectID:     projectID,
		traceExporter: tex,
	}, nil
}

// getTraceExporter is responsible for collecting traces
// and sending them to Cloud Trace via the Stackdriver exporter
func getTraceExporter(projectID string) (traceExporter, error) {
	_, flush, err := texporter.InstallNewPipeline(
		[]texporter.Option{
			texporter.WithProjectID(projectID),
		},
		sdktrace.WithConfig(sdktrace.Config{
			DefaultSampler: sdktrace.AlwaysSample(),
		}),
	)
	if err != nil {
		return traceExporter{}, errors.Wrap(err, "installing pipeline")
	}
	t := global.Tracer(parentSpanName)
	ctx, span := t.Start(context.Background(), parentSpanName)
	cleanup := func() {
		span.End()
		flush()
	}
	return traceExporter{
		parentCtx: ctx,
		cleanup:   cleanup,
		Tracer:    t,
		spans: map[string]trace.Span{
			parentSpanName: span,
		},
	}, nil
}

// getMetricExporter is responsible for collecting one metric (start time)
// and sending it to Cloud Monitoring via the Stackdriver exporter
func getMetricExporter(projectID string) (metricExporter, error) {
	osMethod, err := tag.NewKey("os")
	if err != nil {
		return metricExporter{}, errors.Wrap(err, "new tag key")
	}

	ctx, err := tag.New(context.Background(), tag.Insert(osMethod, runtime.GOOS))
	if err != nil {
		return metricExporter{}, errors.Wrap(err, "new tag")
	}
	latencyS := stats.Float64("repl/start_time", "start time in seconds", "s")
	// Register the view. It is imperative that this step exists,
	// otherwise recorded metrics will be dropped and never exported.
	v := &view.View{
		Name:        customMetricName,
		Measure:     latencyS,
		Aggregation: view.LastValue(),
	}
	if err := view.Register(v); err != nil {
		return metricExporter{}, errors.Wrap(err, "registering view")
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
		// ReportingInterval sets the frequency of reporting metrics
		// to stackdriver backend.
		ReportingInterval: 1 * time.Second,
	})
	if err != nil {
		return metricExporter{}, errors.Wrap(err, "new exporter")
	}
	if err := sd.StartMetricsExporter(); err != nil {
		return metricExporter{}, errors.Wrap(err, "starting metrics exporter")
	}
	now := time.Now()

	cleanup := func() {
		// record the total time this command took
		fmt.Println(time.Since(now).Seconds())
		stats.Record(ctx, latencyS.M(time.Since(now).Seconds()))
		sd.Flush()
		time.Sleep(5 * time.Second)
		sd.StopMetricsExporter()
	}

	return metricExporter{
		cleanup: cleanup,
	}, nil
}
