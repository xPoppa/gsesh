package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	fzf "github.com/junegunn/fzf/src"
)

func exit(code int, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(code)
}

type sessions map[string]int

func main() {
	// This should run in a go routine so when it gets killed it should write things to permanent storage
	// Further should it clean up all orphaned go routines? I don't know for now nope
	// The sigint story by killing the application and writing it to a file is for later This way you can just run the application and then it will write to the storage and stuff
	// Probably have to make a client server model. Especially as I want to run commands and shit and have them be persistent
	keyVal := sessions{}
	for {
		dirs, err := findCmd()
		if err != nil {
			fmt.Printf("Find failed with error: %s\n", err)
			os.Exit(1)
		}
		code, err, out := fzff(dirs)
		fmt.Println(out)
		fmt.Println(code)
		if code == 130 {
			break
		}
		if err != nil {
			exit(code, err)
		}
		go ghostty(out, keyVal)
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

func ghostty(in string, keyVal sessions) error {
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

func findCmd() (*bytes.Buffer, error) {
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

func fzff(in *bytes.Buffer) (int, error, string) {
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
		[]string{"--multi", "--reverse", "--border", "--height=40%"},
	)

	if err != nil {
		exit(fzf.ExitError, err)
	}
	options.Input = inputChan
	options.Output = outputChan

	// Run fzf
	code, err := fzf.Run(options)
	return code, err, output
}
