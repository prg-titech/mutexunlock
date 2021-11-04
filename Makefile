test:
	go test .

run:
	go run cmd/unlockcheck/main.go ./testdata/src/a

build:
	go build -o ./unlockcheck ./cmd/unlockcheck
