// utility/proc.go
package Utility

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/mitchellh/go-ps"
)

// GetProcessIdsByName returns a list of process IDs that match the given name prefix.
func GetProcessIdsByName(name string) ([]int, error) {
	processList, err := ps.Processes()
	if err != nil {
		return nil, errors.New("ps.Processes() failed, are you using windows?")
	}

	pids := make([]int, 0)
	for _, proc := range processList {
		if strings.HasPrefix(proc.Executable(), name) {
			pids = append(pids, proc.Pid())
		}
	}
	return pids, nil
}

// PidExists checks whether a process with the given PID exists.
func PidExists(pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid pid %v", pid)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if runtime.GOOS == "windows" {
		// Todo find a way to test if the process is really running...
		return true, nil
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	if errno, ok := err.(syscall.Errno); ok {
		switch errno {
		case syscall.ESRCH:
			return false, nil
		case syscall.EPERM:
			return true, nil
		}
	}
	return false, err
}

// GetProcessRunningStatus returns the process if it's alive, or error otherwise.
func GetProcessRunningStatus(pid int) (*os.Process, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	if runtime.GOOS == "windows" {
		return proc, nil
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return proc, nil
	}
	if err == syscall.ESRCH {
		return nil, errors.New("process not running")
	}
	return nil, errors.New("process running but query operation not permitted")
}

// KillProcessByName kills all processes that match the given name prefix.
func KillProcessByName(name string) error {
	pids, err := GetProcessIdsByName(name)
	if err != nil {
		return err
	}
	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			log.Println(err)
			continue
		}
		if proc != nil {
			if !strings.HasPrefix(name, "Globular") {
				if err := proc.Kill(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// TerminateProcess sends an interrupt signal to a process by pid.
func TerminateProcess(pid int, exitcode int) error {
	// Windows implementation can use syscall.TerminateProcess (commented in original code)
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(os.Interrupt)
}


// ReadOutput reads line-oriented output from rc and sends it to the output channel.
// It trims trailing CR for CRLF streams and closes both rc and output when finished.
func ReadOutput(output chan string, rc io.ReadCloser) {
	defer func() {
		_ = rc.Close()
		close(output)
	}()

	sc := bufio.NewScanner(rc)
	// Increase buffer in case of long lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024) // up to 1MB lines

	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		// preserve behavior: skip pure empty lines after trim
		if strings.TrimSpace(line) != "" {
			output <- line
		}
	}
	if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
		log.Println("ReadOutput:", err)
	}
}

// RunCmd executes a command in dir with args and streams stdout lines to the console.
// It sends the final error (nil on success) on wait and returns.
// Stdout is streamed; stderr is captured and included in the error on failure.
func RunCmd(name, dir string, args []string, wait chan error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		wait <- err
		return
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Println("run command:", name, args)

	// Start the command before launching readers; if Start fails, we won't block on pipes.
	if err := cmd.Start(); err != nil {
		wait <- fmt.Errorf("%s </br> %w: %s", buildCmdLine(name, args), err, stderr.String())
		return
	}

	// Channel to receive stdout lines and a signal when printing is done
	outCh := make(chan string, 256)
	donePrint := make(chan struct{})

	// Printer goroutine: echo every stdout line with command and pid
	go func() {
		for line := range outCh {
			pid := -1
			if cmd.Process != nil {
				pid = cmd.Process.Pid
			}
			fmt.Println(name+":", pid, line)
		}
		close(donePrint)
	}()

	// Reader goroutine: reads stdout and closes outCh when finished
	go ReadOutput(outCh, stdout)

	// Wait for the command to exit
	err = cmd.Wait()

	// Ensure we finish printing any remaining lines
	<-donePrint

	if err != nil {
		wait <- fmt.Errorf("%s </br> %v: %s", buildCmdLine(name, args), err, strings.TrimSpace(stderr.String()))
		return
	}

	wait <- nil
}

// buildCmdLine formats `name` and `args` into a shell-like string.
func buildCmdLine(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	var b strings.Builder
	b.WriteString(name)
	for _, a := range args {
		b.WriteByte(' ')
		b.WriteString(a)
	}
	return b.String()
}