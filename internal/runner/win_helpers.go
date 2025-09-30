// internal/runer/windows.go
//go:build windows

package runner

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/patnaikankit/Gaze/pkg/logger"
)

// terminateProcess tries to gracefully stop a process by PID
func terminateProcess(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid))
	return cmd.Run()
}

// forceKillProcess forcefully kills a process by PID
func forceKillProcess(pid int) error {
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill failed for PID %d: %v\nOutput: %s", pid, err, string(output))
	}
	return nil
}

// cleanupPort tries to free a port by killing processes and removing lingering connections
func cleanupPort(port string) error {
	logger.Warn("Cleaning up port %s...", port)

	// Find processes using the port
	findCmd := exec.Command("cmd", "/C", fmt.Sprintf("netstat -ano | findstr :%s | findstr LISTENING", port))
	output, err := findCmd.CombinedOutput()

	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\r\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pidStr := fields[4]
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

	// Try cleaning via netsh
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}

	commands := []string{
		fmt.Sprintf("netsh int ipv4 delete tcpconnection localport=%d", portNum),
		fmt.Sprintf("netsh int ipv4 delete tcpconnection 0.0.0.0:%d", portNum),
		fmt.Sprintf("netsh int ipv4 delete tcpconnection 127.0.0.1:%d", portNum),
		fmt.Sprintf("netsh int ipv6 delete tcpconnection localport=%d", portNum),
	}

	for _, cmdStr := range commands {
		_ = exec.Command("cmd", "/C", cmdStr).Run()
	}

	// Verify port is free
	checkCmd := exec.Command("cmd", "/C", fmt.Sprintf("netstat -ano | findstr :%s | findstr LISTENING", port))
	checkOutput, _ := checkCmd.CombinedOutput()

	if len(checkOutput) > 0 {
		logger.Warn("Port %s might still be in use after cleanup attempts", port)
	} else {
		logger.Info("Port %s successfully freed", port)
	}

	return nil
}
