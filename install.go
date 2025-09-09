package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"text/template"

	"github.com/xPoppa/gsesh/server"
)

const (
	INSTALL_PATH = "/.local/bin/"
)

func main() {
	os.Mkdir("build", 0660)
	u, err := user.Current()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Cannot find user home dir: ", err)
	}
	gsesh := homeDir + INSTALL_PATH + "gsesh"
	gseshServer := homeDir + INSTALL_PATH + "gsesh_server"

	cmd := exec.Command("go", "build", "-o", gsesh, "cmd/cli/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println("Failed to run command")
		os.Exit(1)
		return
	}

	cmd = exec.Command("go", "build", "-o", gseshServer, "cmd/server/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println("Failed to run command")
		os.Exit(1)
		return
	}

	tmpl, err := template.ParseFiles("build/systemd.tmpl")
	tmpl.Execute(os.Stdout, struct {
		StorageDir string
		BinDir     string
		User       string
	}{
		StorageDir: server.STORAGE_DIR,
		BinDir:     gseshServer,
		User:       u.Username,
	})
}
