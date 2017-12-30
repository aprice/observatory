package utils

import (
	"bufio"
	"log"
	"os/exec"
	"syscall"
)

// Execute a check and return its exit code.
// Based on http://stackoverflow.com/a/10385867/7426 and
// http://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/.
func Execute(args ...string) (exitCode int, output []byte, err error) {
	exitCode = -1
	log.Printf("Executing %s with %d args: %v", args[0], len(args)-1, args[1:])
	cmd := exec.Command(args[0])
	cmd.Args = args

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Error creating StdoutPipe for Cmd", err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			output = append(output, scanner.Bytes()...)
		}
	}()

	if err = cmd.Start(); err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status: %d", waitStatus.ExitStatus())
				exitCode = waitStatus.ExitStatus()
			} else {
				log.Printf("Unable to coerce to syscall.WaitStatus")
				return
			}
		} else {
			log.Printf("Unable to coerce to exec.ExitError")
			return
		}
	} else if waitStatus, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
		log.Printf("Exit Status: %d", waitStatus.ExitStatus())
		exitCode = waitStatus.ExitStatus()
	} else {
		exitCode = 0
	}

	return
}
