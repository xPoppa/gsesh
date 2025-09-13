package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	db "github.com/xPoppa/gsesh/internal"
)

const (
	SERVER_SOCKET = "/tmp/gsesh.sock"
	STORAGE_DIR   = "/home/poppa/.local/share/gsesh"
)

type Server struct {
	db       *db.DB
	listener net.Listener
}

func Run() error {
	server, err := setup()
	if err != nil {
		return err
	}
	defer func() {
		server.db.Close()
		server.listener.Close()
		err := os.Remove(SERVER_SOCKET)
		if err != nil {
			log.Println("Socket already removed")
		}
	}()
	return server.Serve()
}

func setup() (*Server, error) {
	os.Remove(SERVER_SOCKET)
	err := os.MkdirAll(STORAGE_DIR, 0700)
	if err != nil {
		return nil, err
	}

	store, err := db.NewDB(STORAGE_DIR + "/my.db")
	if err != nil {
		return nil, err
	}

	err = checkPidsAndRemoveInactive(store)
	if err != nil {
		return nil, err
	}

	listen, err := net.Listen("unix", SERVER_SOCKET)
	if err != nil {
		return nil, err
	}
	log.Println("Server listening on socket: ", SERVER_SOCKET)
	return &Server{db: store, listener: listen}, nil
}

// TODO: Add more commands to server, list sessions, kill session
func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error while accepting listeners on: ", SERVER_SOCKET, " with err: ", err)
			return err
		}

		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Error reading message: ", string(message))
			continue
		}

		errChan := make(chan error)
		go ghostty(string(message[:len(message)-1]), s.db, errChan)

		go func() {
			err := <-errChan
			if err != nil {
				log.Println("Making ghostty window failed with err: ", err)
			}
		}()
	}
}

// TODO: Make output a struct. Nicely interface with wmctrl
// TODO: Port wmctrl?
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

func handleExistingSession(key string, pid int) error {
	buff := &bytes.Buffer{}
	log.Println("Already exists: ", key)
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
		log.Printf("HANDLE_EXISTSING: wmctrl output: %s", s)
		if strings.Contains(s, fmt.Sprint(pid)) {

			splittedIter := strings.Split(s, " ")
			goToWindow(splittedIter[0])

			return nil
		}
	}

	log.Printf("Splitted res: %+v", splitted)

	return nil
}

func ghostty(key string, db *db.DB, errChan chan error) {
	res, err := db.GetPid(key)
	if err != nil {
		errChan <- err
		return
	}
	if res.Exists {
		errChan <- handleExistingSession(key, res.Pid)
		return
	}

	cmd := exec.Command("ghostty", "--working-directory="+key)
	err = cmd.Start()
	if err != nil {
		errChan <- err
		return
	}
	log.Printf("Insert: %s in session\n", key)
	err = db.Insert(key, cmd.Process.Pid)
	if err != nil {
		errChan <- err
		return
	}

	cmd.Wait()
	if cmd.ProcessState.ExitCode() != -1 {
		log.Printf("Deleting from session: %s\n", key)
		db.Delete(key)
		return
	}
	errChan <- err
	return
}

func checkPidsAndRemoveInactive(db *db.DB) error {
	pids, err := db.ReturnPids()
	if err != nil {
		return err
	}
	for _, pid := range pids {
		proc, err := os.FindProcess(pid.Pid)
		if err != nil {
			return err
		}
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			err := db.Delete(pid.Key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
