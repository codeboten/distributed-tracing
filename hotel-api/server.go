package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

type Hotel struct {
	Name string
}

func thisWasHereAndWeDontKnowWhy(ctx context.Context, city string) {
	_, span := trace.StartSpan(ctx, "thisWasHereAndWeDontKnowWhy")
	span.AddAttributes(trace.StringAttribute("city", city))
	defer span.End()
}

func getHotels(ctx context.Context, city string) []Hotel {
	_, span := trace.StartSpan(ctx, "getHotels")
	span.AddAttributes(trace.StringAttribute("city", city))

	defer span.End()
	if city == "Vegas" {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 250)
			thisWasHereAndWeDontKnowWhy(ctx, city)
		}

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
	ctx, span := trace.StartSpan(r.Context(), "handleHotelsRequest")

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
		CollectorEndpoint: os.Getenv("TRACING_URL"),
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
	log.Fatal(http.ListenAndServe(":8081", &ochttp.Handler{}))
}
