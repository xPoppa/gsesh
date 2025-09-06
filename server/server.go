package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
)

var STORAGE_DIR = "/home/poppa/.local/share/gsesh"

const (
	SERVER_SOCKET = "/tmp/gsesh.sock"
	BUCKET_NAME   = "sessions"
)

type sessions map[string]int

// This is all server code

func Run() {
	// This should run in a go routine so when it gets killed it should write things to permanent storage
	// Further should it clean up all orphaned go routines? I don't know for now nope
	// The sigint story by killing the application and writing it to a file is for later This way you can just run the application and then it will write to the storage and stuff
	// Probably have to make a client server model. Especially as I want to run commands and shit and have them be persistent
	err := os.MkdirAll(STORAGE_DIR, 0700)
	if err != nil {
		log.Fatal(err)
	}

	db, err := bolt.Open(STORAGE_DIR+"/my.db", 0660, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_NAME))
		if err != nil {
			return err
		}
		return nil
	})
	fmt.Println("Server listening on socket: ", SERVER_SOCKET)
	listen, err := net.Listen("unix", SERVER_SOCKET)
	if err != nil {
		log.Fatal("Cannot open socket: ", SERVER_SOCKET, " with err: ", err)
	}
	defer listen.Close()
	defer func() {
		os.Remove(SERVER_SOCKET)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(SERVER_SOCKET)
		db.Close()
		os.Exit(1)
	}()

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

		//fmt.Print("Message Received:", string(message[:len(message)-2]))
		go ghostty(string(message[:len(message)-1]), db)
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
func ghostty(in string, db *bolt.DB) error {
	fmt.Println("The in string", in)

	var pid int
	var exists bool
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_NAME))
		bpid := b.Get([]byte(in))
		if bpid != nil {
			exists = true
			p, err := strconv.Atoi(string(bpid))
			if err != nil {
				return err
			}
			pid = p
			return nil
		}
		return nil
	})
	if exists {
		return handleExistingSession(in, pid)
	}

	cmd := exec.Command("ghostty", "--working-directory="+in)
	err := cmd.Start()
	if err != nil {
		return err
	}
	fmt.Printf("Insert: %s in session\n", in)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_NAME))
		return b.Put([]byte(in), []byte(strconv.Itoa(cmd.Process.Pid)))

	})

	cmd.Wait()
	if cmd.ProcessState.ExitCode() != -1 {
		fmt.Printf("Deleting from session: %s\n", in)
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(BUCKET_NAME))
			return b.Delete([]byte(in))
		})
	}
	return nil
}
