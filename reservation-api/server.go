package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	"contrib.go.opencensus.io/exporter/jaeger"
	honeycomb "github.com/honeycombio/opencensus-exporter/honeycomb"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

type WeatherReport struct {
	Temperature int
	Condition   string
	Error       string
}

type Trip struct {
	City     string        `json:"city"`
	Forecast WeatherReport `json:"forecast"`
	Hotels   []Hotel       `json:"hotels"`
	Estimate float64       `json:"estimate"`
}

type Hotel struct {
	Name string
}

func estimateCost(city string) float64 {
	totalCost := 923923.0
	// no one gets what this does but its important
	return totalCost
}

func getWeather(ctx context.Context, city string) WeatherReport {
	localContext, span := trace.StartSpan(ctx, "getWeather")
	defer span.End()

	client := &http.Client{Transport: &ochttp.Transport{}}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("WEATHER_URL"), city), nil)
	if err != nil {
		fmt.Println("Got an error yall")
		fmt.Printf("%v", err)
		return WeatherReport{}
	}

	req = req.WithContext(localContext)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Got an error yall")
		fmt.Printf("%v", err)
		return WeatherReport{}
	}

	var report WeatherReport
	json.NewDecoder(res.Body).Decode(&report)

	return report
}

func getHotels(ctx context.Context, city string) []Hotel {
	localContext, span := trace.StartSpan(ctx, "getHotels")
	defer span.End()

	client := &http.Client{Transport: &ochttp.Transport{}}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("HOTELS_URL"), city), nil)
	if err != nil {
		fmt.Println("Got an error yall")
		fmt.Printf("%v", err)
		return []Hotel{}
	}
	req = req.WithContext(localContext)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Got an error yall")
		fmt.Printf("%v", err)
		return []Hotel{}
	}

	defer res.Body.Close()

	var hotels []Hotel
	json.NewDecoder(res.Body).Decode(&hotels)

	return hotels
}

func handleReservationRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "handleReservationRequest")
	defer span.End()
	// lookup travel advisories
	// get a quote for the trip
	// find a list of hotels
	// get weather report for the dates
	city := r.FormValue("city")
	if len(city) == 0 {
		span.SetStatus(trace.Status{Code: 13, Message: "Invalid city specified"})
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "404 - City not found!"}`))
		return
	}
	trip := Trip{
		City:     city,
		Forecast: getWeather(ctx, city),
		Hotels:   getHotels(ctx, city),
	}
	json.NewEncoder(w).Encode(trip)
}

func main() {
	// Register the Jaeger exporter to be able to retrieve
	// the collected spans.
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: os.Getenv("TRACING_URL"),
		Process: jaeger.Process{
			ServiceName: "reservation-api",
		},
	})

	honeycombExporter := honeycomb.NewExporter(os.Getenv("HONEYCOMB_KEY"), os.Getenv("HONEYCOMB_DATASET"))

	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}
	trace.RegisterExporter(exporter)
	trace.RegisterExporter(honeycombExporter)

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	http.HandleFunc("/v1/reservation", handleReservationRequest)
	log.Fatal(http.ListenAndServe(":8080", &ochttp.Handler{}))
}
