package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
)

type Hotel struct {
	Name string
}

func getHotels(ctx context.Context, city string) []Hotel {
	_, span := trace.StartSpan(ctx, "getHotels")

	defer span.End()
	if city == "Vegas" {
		time.Sleep(time.Second * 5)
		return []Hotel{
			Hotel{Name: "Bellagio"},
			Hotel{Name: "MGM Grand"},
			Hotel{Name: "Hilton"},
			Hotel{Name: "Holiday Inn"},
		}
	}

	return []Hotel{
		Hotel{Name: "Westin"},
		Hotel{Name: "Fairmont"},
	}
}

func handleHotelsRequest(w http.ResponseWriter, r *http.Request) {
	f := &b3.HTTPFormat{}
	var span *trace.Span
	var ctx context.Context
	spanCtx, ok := f.SpanContextFromRequest(r)
	time.Sleep(time.Millisecond * 90)

	if ok {
		ctx, span = trace.StartSpanWithRemoteParent(context.Background(), "handleHotelsRequest", spanCtx)
	} else {
		ctx, span = trace.StartSpan(context.Background(), "handleHotelsRequest")
	}

	defer span.End()
	city := r.FormValue("city")
	span.AddAttributes(trace.StringAttribute("city", city))
	if city == "ErrorLand" {
		span.SetStatus(trace.Status{Code: 13, Message: "ErrorLand is not a real place, error occurred"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	} else {
		json.NewEncoder(w).Encode(getHotels(ctx, city))
	}
}

func main() {
	// Register the Jaeger exporter to be able to retrieve
	// the collected spans.
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint: "http://localhost:14268",
		Process: jaeger.Process{
			ServiceName: "hotel-api",
		},
	})

	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}
	trace.RegisterExporter(exporter)

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	http.HandleFunc("/v1/hotels", handleHotelsRequest)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
