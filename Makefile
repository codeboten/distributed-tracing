build:
	go build -o this-will-totally-work

build-hotel-api:
	cd hotel-api && go build

build-weather-api:
	cd weather-api && go build

demo: build build-hotel-api build-weather-api
	docker run -d --name jaeger \
	-e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
	-p 5775:5775/udp \
	-p 6831:6831/udp \
	-p 6832:6832/udp \
	-p 5778:5778 \
	-p 16686:16686 \
	-p 14268:14268 \
	-p 9411:9411 \
	jaegertracing/all-in-one:1.8
