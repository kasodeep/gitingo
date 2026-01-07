APP := gitingo

build:
	@go build -o bin/$(APP)

run: build
	@./bin/$(APP) $(ARGS)

test:
	@go test -v ./...