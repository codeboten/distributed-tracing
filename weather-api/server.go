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

type WeatherReport struct {
	Temperature float64
	Condition   string
}

func getWeather(ctx context.Context, city string) WeatherReport {
	_, span := trace.StartSpan(ctx, "getWeather")
	time.Sleep(time.Millisecond * 50)

	defer span.End()

	return WeatherReport{
		Temperature: -9.1,
		Condition:   "sunny",
	}
}

func handleWeatherRequest(w http.ResponseWriter, r *http.Request) {
	f := &b3.HTTPFormat{}
	var span *trace.Span
	var ctx context.Context
	spanCtx, ok := f.SpanContextFromRequest(r)

	if ok {
		ctx, span = trace.StartSpanWithRemoteParent(context.Background(), "handleWeatherRequest", spanCtx)
	} else {
		ctx, span = trace.StartSpan(context.Background(), "handleWeatherRequest")
	}

	defer span.End()
	// lookup travel advisories
	// get a quote for the trip
	// find a list of hotels
	// get weather report for the dates
	// city := "Whistler"
	city := r.FormValue("city")
	span.AddAttributes(trace.StringAttribute("city", city))
	json.NewEncoder(w).Encode(getWeather(ctx, city))
}

func main() {
	// Register the Jaeger exporter to be able to retrieve
	// the collected spans.
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint: "http://localhost:14268",
		Process: jaeger.Process{
			ServiceName: "weather-api",
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
	http.HandleFunc("/v1/weather", handleWeatherRequest)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
