package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/xPoppa/gsesh/client"
)

type sessions map[string]int

const (
	SERVER_SOCKET = "/tmp/gsesh.sock"
)

// This is all server code

func handleConn(c net.Conn) {
	defer c.Close()
	for {
		_, err := io.WriteString(c, "Hi from server\n")
		if err != nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func main() {
	// This should run in a go routine so when it gets killed it should write things to permanent storage
	// Further should it clean up all orphaned go routines? I don't know for now nope
	// The sigint story by killing the application and writing it to a file is for later This way you can just run the application and then it will write to the storage and stuff
	// Probably have to make a client server model. Especially as I want to run commands and shit and have them be persistent
	keyVal := sessions{}
	fmt.Println("Server listening on socket: ", SERVER_SOCKET)
	listen, err := net.Listen("unix", SERVER_SOCKET)
	if err != nil {
		log.Fatal("Cannot open socket: ", SERVER_SOCKET, " with err: ", err)
	}
	go client.Run()
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("Error while accepting listeners on: ", SERVER_SOCKET, " with err: ", err)
			continue
		}

		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print("Message Received:", string(message[:len(message)-2]))
		go ghostty(string(message[:len(message)-2]), keyVal)
	}
}

// server code
func goToWindow(w string) error {
	cmd := exec.Command("wmctrl", "-ia", w)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

// server code
func handleExistingSession(in string, pid int) error {
	buff := &bytes.Buffer{}
	fmt.Println("Already exists: ", in)
	cmd := exec.Command("wmctrl", "-lp")
	cmd.Stdout = buff

	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	res := buff.String()
	splitted := strings.Split(res, "\n")
	for _, s := range splitted {
		if strings.Contains(s, fmt.Sprint(pid)) {

			splittedIter := strings.Split(s, " ")
			goToWindow(splittedIter[0])

			return nil
		}
	}

	fmt.Printf("Splitted res: %s", splitted)

	return nil
}

// This is server code
func ghostty(in string, keyVal sessions) error {
	fmt.Println("The in string", in)
	if pid, ok := keyVal[in]; ok {
		return handleExistingSession(in, pid)
	}
	cmd := exec.Command("ghostty", "--working-directory="+in)
	err := cmd.Start()
	if err != nil {
		return err
	}
	fmt.Printf("Insert: %s in session\n", in)
	keyVal[in] = cmd.Process.Pid
	cmd.Wait()

	if cmd.ProcessState.ExitCode() != -1 {
		fmt.Printf("Deleting from session: %s\n", in)
		delete(keyVal, in)
	}
	return nil
}
