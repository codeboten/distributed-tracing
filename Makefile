build-all: build build-hotel-api build-weather-api

build-reservation-api:
	cd reservation-api && go build

build-hotel-api:
	cd hotel-api && go build

build-weather-api:
	cd weather-api && go build

clean-reservation-api:
	cd reservation-api && rm -f ./reservation-api

clean-hotel-api:
	cd hotel-api && rm -f ./hotel-api

clean-weather-api:
	cd weather-api && rm -f ./weather-api

demo:
	docker-compose build
	docker-compose up

clean: clean-reservation-api clean-hotel-api clean-weather-api
	@docker rm -f jaeger | true