// internal/runner/unix.go
//go:build darwin || linux

package runner

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/patnaikankit/Gaze/pkg/logger"
)

// terminateProcess sends SIGTERM to a PID
func terminateProcess(pid int) error {
	cmd := exec.Command("kill", "-15", strconv.Itoa(pid))
	return cmd.Run()
}

// forceKillProcess sends SIGKILL to a PID
func forceKillProcess(pid int) error {
	cmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kill -9 failed for PID %d: %v\nOutput: %s", pid, err, string(output))
	}
	return nil
}

// cleanupPort finds processes listening on the port and kills them
func cleanupPort(port string) error {
	logger.Warn("Cleaning up port %s...", port)

	// Find processes with lsof
	findCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%s", port), "-sTCP:LISTEN", "-n", "-P")
	output, err := findCmd.CombinedOutput()
	if err != nil {
		logger.Debug("lsof did not find processes for port %s: %v", port, err)
	} else if len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" || strings.HasPrefix(line, "COMMAND") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pidStr := fields[1]
				pid, err := strconv.Atoi(pidStr)
				if err != nil {
					logger.Warn("Invalid PID: %s", pidStr)
					continue
				}

				logger.Info("Killing process %d using port %s", pid, port)
				if err := forceKillProcess(pid); err != nil {
					logger.Error("Failed to kill PID %d: %v", pid, err)
				} else {
					logger.Debug("Successfully killed process with PID %d", pid)
				}
			}
		}
	}

	// Verify port is free
	checkCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%s", port), "-sTCP:LISTEN")
	checkOutput, _ := checkCmd.CombinedOutput()

	if len(checkOutput) > 0 {
		logger.Warn("Port %s might still be in use after cleanup attempts", port)
	} else {
		logger.Info("Port %s successfully freed", port)
	}

	return nil
}
