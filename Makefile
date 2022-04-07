test:
	go test ./...

build:
	go build -o ./mutexunlock ./cmd/mutexunlock

install:
	go install ./cmd/mutexunlock

run:
	go run ./cmd/mutexunlock/main.go ./testdata/src/a

clean:
	rm -rf ./mutexunlock

play-clean:
	@rm -rf ./_playground

play-init:
	@make play-clean
	@mkdir _playground
	@cp -r ./testdata/src/a/* ./_playground/

play:
	@make build
	@make play-init
	./mutexunlock -fix ./_playground

debug:
	@make build
	@make play-init
	VERBOSE_LEVEL=2 ./mutexunlock ./_playground
