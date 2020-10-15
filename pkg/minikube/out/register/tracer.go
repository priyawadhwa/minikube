package register

import (
	"context"
	"fmt"

	"github.com/priyawadhwa/tracer-minikube/pkg/tracer"
	"go.opentelemetry.io/otel/api/trace"
)

type mktracer struct {
	t trace.Tracer
}

var (
	t mktracer
)

// InitializeTracer intializes the tracer
func InitializeTracer(t string) error {
	switch t {
	case "gcp":
		return initGCPTracer()
	}
	return fmt.Errorf("%s is not a valid tracer", t)
}

func initGCPTracer() error {
	// projectID := os.Getenv("MINIKUBE_TRACER_GCP_PROJECT_ID")
	// if projectID == "" {
	// 	return fmt.Errorf("gcp tracer requires a valid GCP project id set via the MINIKUBE_TRACER_GCP_PROJECT_ID env variable")
	// }
	// exporter, err := texporter.NewExporter(texporter.WithProjectID(projectID))

	// if err != nil {
	// 	return errors.Wrap(err, "getting exporter")
	// }
	// tp, err := sdktrace.NewProvider(sdktrace.WithSyncer(exporter))
	// if err != nil {
	// 	return errors.Wrap(err, "new provider")
	// }
	// tp.ApplyConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()})
	// global.SetTraceProvider(tp)
	// t = tracer{
	// 	t: global.TraceProvider().Tracer("container-tools"),
	// }

	_, s := tracer.StartSpan(context.Background(), "span")
	defer s.End()

	return nil
}
