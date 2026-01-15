.PHONY: run test

run:
	go run ./cmd/server

test:
	go test -count=1 ./... -v