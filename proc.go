// utility/proc.go
package Utility

import (
	"errors"
	"fmt"
	"log"
	"os"
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

