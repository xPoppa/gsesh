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
		os.Remove(SERVER_SOCKET)
	}()
	return server.Serve()
}

func setup() (*Server, error) {
	err := os.Remove(SERVER_SOCKET)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(STORAGE_DIR, 0700)
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
	fmt.Println("Server listening on socket: ", SERVER_SOCKET)
	return &Server{db: store, listener: listen}, nil
}

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
	fmt.Println("Already exists: ", key)
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

	fmt.Printf("Splitted res: %+v", splitted)

	return nil
}

func ghostty(key string, db *db.DB, errChan chan error) {
	fmt.Println("The in string", key)

	res, err := db.GetPid(key)
	if err != nil {
		errChan <- err
	}
	if res.Exists {
		errChan <- handleExistingSession(key, res.Pid)
	}

	cmd := exec.Command("ghostty", "--working-directory="+key)
	err = cmd.Start()
	if err != nil {
		errChan <- err
	}
	fmt.Printf("Insert: %s in session\n", key)

	cmd.Wait()
	if cmd.ProcessState.ExitCode() != -1 {
		fmt.Printf("Deleting from session: %s\n", key)
		db.Delete(key)
	}
	errChan <- err
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
