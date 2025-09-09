package client

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	fzf "github.com/junegunn/fzf/src"
)

const (
	SERVER_SOCKET = "/tmp/gsesh.sock"
)

func Run() {
	buf, err := FindCmd()
	if err != nil {
		log.Fatal("Something went wrong with find: ", err)
	}
	FuzzyFind(buf)
}

// client code
// TODO: Abstract command if needed?
// TODO:  take directories as input
// TODO: List sessions
// TODO:
func FindCmd() (*bytes.Buffer, error) {
	homeDir, _ := os.UserHomeDir()
	findCmd := exec.Command(
		"find",
		"-L",
		homeDir+"/boot.dev",
		homeDir+"/go",
		homeDir+"/haskell",
		homeDir+"/frontendmasters",
		homeDir+"/Projects",
		homeDir+"/natalie",
		homeDir+"/",
		"-mindepth", "1",
		"-maxdepth", "1",
		"-type", "d")
	buff := bytes.NewBuffer([]byte{})
	findCmd.Stdout = buff
	err := findCmd.Start()
	if err != nil {
		return buff, err
	}
	findCmd.Wait()
	return buff, nil
}

// This is client code
func FuzzyFind(in *bytes.Buffer) {
	inputChan := make(chan string)
	go func() {
		for s := range strings.SplitSeq(in.String(), "\n") {
			inputChan <- s
		}
		close(inputChan)
	}()

	output := ""

	outputChan := make(chan string)
	go func() {
		for s := range outputChan {
			output = s
		}
	}()

	options, err := fzf.ParseOptions(
		true, // whether to load defaults ($FZF_DEFAULT_OPTS_FILE and $FZF_DEFAULT_OPTS)
		[]string{"--multi", "--reverse", "--border", "--height=80%"},
	)

	if err != nil {
		exit(fzf.ExitError, err)
	}
	options.Input = inputChan
	options.Output = outputChan

	// Run fzf
	code, err := fzf.Run(options)
	// send output
	fmt.Println(output)
	if output == "" {
		exit(code, err)
	}
	sendToConn(output)
	exit(code, err)
}

func sendToConn(output string) {
	conn, err := net.Dial("unix", SERVER_SOCKET)
	defer conn.Close()
	if err != nil {
		log.Fatal("Cannot dial socket with err: ", err)
	}
	fmt.Fprint(conn, output+"\n")
}

func exit(code int, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(code)
}
