package main

import (
	"encoding/json"
	"log"

	"net/http"
	_ "net/http/pprof"
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
	return totalCost
}

func getWeather(city string) WeatherReport {
	return WeatherReport{
		Temperature: -9.1,
		Condition:   "sunny",
	}
}

func getHotels(city string) []Hotel {
	return []Hotel{
		Hotel{Name: "Westin"},
		Hotel{Name: "Fairmont"},
	}
}

func handleReservationRequest(w http.ResponseWriter, r *http.Request) {
	// lookup travel advisories
	// get a quote for the trip
	// find a list of hotels
	// get weather report for the dates
	city := "Whistler"
	trip := Trip{
		City:     city,
		Forecast: getWeather(city),
		Hotels:   getHotels(city),
	}
	json.NewEncoder(w).Encode(trip)
}

func main() {
	http.HandleFunc("/v1/reservation", handleReservationRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
	//fmt.Println("XXXXXXX MADE IT HERE")
	// ships := getSpaceships()
	// refuelFleet(ships)
	// launchFleet(ships)
	// fmt.Println("XXXXXXX NOW WE'RE COOKING WITH GAS")
}
