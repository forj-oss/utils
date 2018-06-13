package utils

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/kvz/logstreamer"
)

// RunCmd is a simple function to call a shell command and display to stdout
// stdout is displayed as is when it arrives, while stderr is displayed in Red, line per line.
func RunCmd(command string, args ...string) int {
	cmd := exec.Command(command, args...)
	gotrace.Trace("RUNNING: %s %s", command, strings.Join(args, " "))
	if gotrace.IsDebugMode() {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		// Setup a streamer that we'll pipe cmd.Stdout to
		logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
		defer logStreamerOut.Close()
		// Setup a streamer that we'll pipe cmd.Stderr to.
		// We want to record/buffer anything that's written to this (3rd argument true)
		logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", true)
		defer logStreamerErr.Close()

		cmd.Stderr = logStreamerErr
		cmd.Stdout = logStreamerOut

		// Reset any error we recorded
		logStreamerErr.FlushRecord()
	}

	// Execute command
	if err := cmd.Start(); err != nil {
		gotrace.Error("ERROR could not spawn command. %s.", err)
		return 255
	}

	if err := cmd.Wait(); err != nil {
		gotrace.Error("\nERROR: wait failure - %s: %s.", command, err)
		return 1
	}
	gotrace.Trace("Command done")
	if status := cmd.ProcessState.Sys().(syscall.WaitStatus); status.ExitStatus() != 0 {
		gotrace.Error("\n%s ERROR: Unable to get process status - %s: %s", command, cmd.ProcessState.String())
		return status.ExitStatus()
	}
	return 0
}

// RunCmdOutput run a command and return the standard output as result.
// stderr is displayed in Red, line per line.
func RunCmdOutput(command string, args ...string) (string, int) {
	cmd := exec.Command(command, args...)
	if gotrace.IsDebugMode() {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		// Setup a streamer that we'll pipe cmd.Stderr to.
		// We want to record/buffer anything that's written to this (3rd argument true)
		logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", true)
		defer logStreamerErr.Close()

		cmd.Stderr = logStreamerErr

		// Reset any error we recorded
		logStreamerErr.FlushRecord()
	}

	gotrace.Trace("RUNNING: %s %s", command, strings.Join(args, " "))

	var scan *bufio.Scanner

	if r, err := cmd.StdoutPipe(); err != nil {
		gotrace.Error("ERROR. Could not pipe the command. %s.", err)
		return "", 255
	} else {
		scan = bufio.NewScanner(r)
	}

	// Execute command
	if err := cmd.Start(); err != nil {
		gotrace.Error("ERROR could not spawn command. %s.", err)
		return "", 255
	}

	if err := cmd.Wait(); err != nil {
		gotrace.Error("\nERROR: wait failure - %s: %s.", command, err)
		return "", 1
	}
	gotrace.Trace("Command done")
	if status := cmd.ProcessState.Sys().(syscall.WaitStatus); status.ExitStatus() != 0 {
		gotrace.Error("\n%s ERROR: Unable to get process status - %s: %s", command, cmd.ProcessState.String())
		return scan.Text(), status.ExitStatus()
	}
	return scan.Text(), 0
}
