all: b

b:
	CGO_ENABLED=0 go build -o bin/influxdb-gateway  main.go

install:
	go install  .
