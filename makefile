APP_NAME:=Go-Git

build:
		@go build -ldflags "-s -w" -race -o ./bin/$(APP_NAME) ./

run: build
		@./bin/$(APP_NAME)

test:
		@go test ./... -v