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
