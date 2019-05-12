package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
)

type WeatherReport struct {
	Temperature float64
	Condition   string
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
	_, span := trace.StartSpan(ctx, "getWeather")
	defer span.End()
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8082/v1/weather?city=%s", city), nil)

	// SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool)
	f := &b3.HTTPFormat{}
	f.SpanContextToRequest(span.SpanContext(), req)
	client := &http.Client{}
	client.Do(req)
	return WeatherReport{
		Temperature: -9.1,
		Condition:   "sunny",
	}
}

func getHotels(ctx context.Context, city string) []Hotel {
	_, span := trace.StartSpan(ctx, "getHotels")

	defer span.End()
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8081/v1/hotels?city=%s", city), nil)

	// SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool)
	f := &b3.HTTPFormat{}
	f.SpanContextToRequest(span.SpanContext(), req)
	client := &http.Client{}
	res, _ := client.Do(req)
	defer res.Body.Close()

	var hotels []Hotel
	json.NewDecoder(res.Body).Decode(&hotels)

	return hotels
}

func handleReservationRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(context.Background(), "handleReservationRequest")
	defer span.End()
	// lookup travel advisories
	// get a quote for the trip
	// find a list of hotels
	// get weather report for the dates
	city := r.FormValue("city")
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
		Endpoint: "http://localhost:14268",
		Process: jaeger.Process{
			ServiceName: "trace-demo",
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

	http.HandleFunc("/v1/reservation", handleReservationRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
	//fmt.Println("XXXXXXX MADE IT HERE")
	// ships := getSpaceships()
	// refuelFleet(ships)
	// launchFleet(ships)
	// fmt.Println("XXXXXXX NOW WE'RE COOKING WITH GAS")
}
