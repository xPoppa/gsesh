package main

import (
	"log"
	"os"
	"os/exec"
)

const (
	INSTALL_PATH = "/.local/bin/"
)

func main() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Cannot find user home dir: ", err)
	}

	cmd := exec.Command("go build", "-o", homeDir+INSTALL_PATH+"gsesh", "cmd/cli/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println("Failed to run command")
		os.Exit(1)
		return
	}

	cmd = exec.Command("go build", "-o", homeDir+INSTALL_PATH+"gsesh_server", "cmd/server/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println("Failed to run command")
		os.Exit(1)
		return
	}
}
