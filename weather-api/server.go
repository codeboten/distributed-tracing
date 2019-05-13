package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/jaeger"
	owm "github.com/briandowns/openweathermap"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
)

type WeatherReport struct {
	Temperature int
	Condition   string
	Error       string
}

func getErrorMessage(ctx context.Context, err error) string {
	_, span := trace.StartSpan(ctx, "getWeather")
	span.AddAttributes(trace.StringAttribute("error", err.Error()))
	defer span.End()
	return err.Error()
}

func getWeather(ctx context.Context, city string) WeatherReport {
	var apiKey = os.Getenv("OWM_API_KEY")
	_, span := trace.StartSpan(ctx, "getWeather")
	defer span.End()

	w, err := owm.NewCurrent("C", "en", apiKey)

	if err != nil {
		return WeatherReport{
			Temperature: 0,
			Condition:   "unavailable",
			Error:       getErrorMessage(ctx, err),
		}
	}

	w.CurrentByName(city)
	if len(w.Weather) == 0 {
		return WeatherReport{
			Temperature: 0,
			Condition:   "unavailable",
			Error:       getErrorMessage(ctx, errors.New("City not found")),
		}
	}

	span.AddAttributes(trace.StringAttribute("weather", w.Weather[0].Description))
	return WeatherReport{
		Temperature: w.Dt,
		Condition:   w.Weather[0].Description,
	}
}

func handleWeatherRequest(w http.ResponseWriter, r *http.Request) {
	f := &b3.HTTPFormat{}
	var span *trace.Span
	var ctx context.Context
	spanCtx, ok := f.SpanContextFromRequest(r)

	if ok {
		ctx, span = trace.StartSpanWithRemoteParent(r.Context(), "handleWeatherRequest", spanCtx)
	} else {
		ctx, span = trace.StartSpan(r.Context(), "handleWeatherRequest")
	}

	defer span.End()
	city := r.FormValue("city")
	span.AddAttributes(trace.StringAttribute("city", city))
	json.NewEncoder(w).Encode(getWeather(ctx, city))
}

func main() {
	// Register the Jaeger exporter to be able to retrieve
	// the collected spans.
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: os.Getenv("TRACING_URL"),
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
	log.Fatal(http.ListenAndServe(":8082", &ochttp.Handler{}))
}
