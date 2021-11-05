test:
	go test .

run:
	go run ./cmd/unlockcheck/main.go ./testdata/src/a

build:
	go build -o ./unlockcheck ./cmd/unlockcheck

clean:
	rm -rf ./unlockcheck

play-clean:
	@rm -rf ./_playground

play-init:
	@make play-clean
	@mkdir _playground
	@cp -r ./testdata/src/a/* ./_playground/

play:
	@make build
	@make play-init
	./unlockcheck -fix ./_playground
