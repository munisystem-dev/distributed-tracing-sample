package main

import (
	"context"
	"log"
	"net/http"
	"path"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func main() {
	exporter := &NopeExporter{}
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(0.5)})

	http.HandleFunc("/a", func(w http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "/a")
		defer span.End()
		sc := span.SpanContext()
		log.Printf("receive /a: sampled %t, trace_id %s, span_id %s", sc.IsSampled(), sc.TraceID, sc.SpanID)
		resp, err := get(req.Context(), "b")
		if err != nil {
			log.Println(err)
		} else {
			resp.Body.Close()
		}
	})
	http.HandleFunc("/b", func(w http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "/a")
		defer span.End()
		sc := span.SpanContext()
		log.Printf("receive /b: sampled %t, trace_id %s, span_id %s", sc.IsSampled(), sc.TraceID, sc.SpanID)
		resp, err := get(req.Context(), "c")
		if err != nil {
			log.Println(err)
		} else {
			resp.Body.Close()
		}
	})
	http.HandleFunc("/c", func(w http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "/a")
		defer span.End()
		sc := span.SpanContext()
		log.Printf("receive /c: sampled %t, trace_id %s, span_id %s", sc.IsSampled(), sc.TraceID, sc.SpanID)
	})

	log.Fatal(http.ListenAndServe(":50030", &ochttp.Handler{}))
}

func get(ctx context.Context, endpoint string) (*http.Response, error) {
	client := &http.Client{Transport: &ochttp.Transport{}}
	r, _ := http.NewRequest("GET", "http://localhost:50030"+path.Join("/", endpoint), nil)
	r = r.WithContext(ctx)
	return client.Do(r)
}

type NopeExporter struct{}

func (e *NopeExporter) ExportView(vd *view.Data) {}

func (e *NopeExporter) ExportSpan(vd *trace.SpanData) {}
