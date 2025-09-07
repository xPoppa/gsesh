server:
  go build -o serv ./cmd/server/main.go

cli:
  go build -o cli ./cmd/cli/main.go

build: server cli

run-server: server
  ./serv

run: cli
  ./cli

default: run

install:
  go run install.go
