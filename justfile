server:
  go build -o serv ./cmd/server/main.go

cli:
  go build -o cli ./cmd/cli/main.go

run-server: server
  ./serv

run: cli
  ./cli

default: run
