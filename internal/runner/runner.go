package runner

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/patnaikankit/Gaze/internal/config"
	"github.com/patnaikankit/Gaze/pkg/logger"
)

type Runner struct {
	cfg     *config.Config
	started time.Time
	cmd     *exec.Cmd
}

func New(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

// Run builds and executes the target binary
func (r *Runner) Run() error {
	// stop old process if running
	if r.isRunning() {
		r.Stop()
		r.cleanupIfWindows()
	}

	args := strings.Fields(r.cfg.RunCMD)
	if len(args) == 0 {
		logger.Error("invalid run command: %q", r.cfg.RunCMD)
		return nil
	}

	r.cmd = exec.Command(args[0], args[1:]...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	r.cmd.Env = append(os.Environ(), r.cmd.Env...)

	if err := r.cmd.Start(); err != nil {
		logger.Error("failed to start application: %v", err)
		return err
	}

	r.started = time.Now()
	logger.Info("started application (PID: %d)", r.cmd.Process.Pid)

	go r.monitor()
	return nil
}

// Stop tries to terminate the process gracefully
func (r *Runner) Stop() {
	if !r.isRunning() {
		return
	}

	pid := r.cmd.Process.Pid

	// If already exited, skip termination logic
	if r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		logger.Info("process (PID: %d) already exited, nothing to stop", pid)
		r.forceCleanup()
		r.cmd = nil
		return
	}

	logger.Debug("stopping application (PID: %d)...", pid)

	err := r.terminate()
	if err != nil {
		logger.Warn("failed to terminate process (PID: %d): %v", pid, err)
		r.forceCleanup()
		return
	}

	done := make(chan struct{})
	go func() {
		_ = r.cmd.Wait() // don't care about error here
		close(done)
	}()

	select {
	case <-done:
		logger.Info("stopped application (PID: %d) after %v", pid, time.Since(r.started))
	case <-time.After(500 * time.Millisecond):
		r.forceKill(pid)
	}

	r.cmd = nil
}

func (r *Runner) isRunning() bool {
	return r.cmd != nil && r.cmd.Process != nil
}

func (r *Runner) monitor() {
	if err := r.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == -1 {
			logger.Warn("application terminated after %v", time.Since(r.started))
		} else {
			logger.Warn("application exited with error: %v after %v", err, time.Since(r.started))
		}
	} else {
		logger.Info("application exited normally after %v", time.Since(r.started))
	}
}

// --- platform helpers ---

func (r *Runner) terminate() error {
	if r.cfg.IsWindows {
		return terminateProcess(r.cmd.Process.Pid)
	}
	return r.cmd.Process.Signal(syscall.SIGTERM)
}

func (r *Runner) forceKill(pid int) {
	var err error
	if r.cfg.IsWindows {
		err = forceKillProcess(pid)
	} else {
		err = r.cmd.Process.Kill()
	}
	if err != nil {
		logger.Error("failed to kill process (PID: %d): %v", pid, err)
	} else {
		logger.Warn("forcibly killed application (PID: %d) after %v", pid, time.Since(r.started))
	}
	r.forceCleanup()
}

func (r *Runner) cleanupIfWindows() {
	if r.cfg.IsWindows && r.cfg.Port != "" {
		time.Sleep(300 * time.Millisecond)
		if err := cleanupPort(r.cfg.Port); err != nil {
			logger.Warn("port cleanup issue: %v", err)
		}
	}
}

func (r *Runner) forceCleanup() {
	if r.cfg.IsWindows && r.cfg.Port != "" {
		time.Sleep(300 * time.Millisecond)
		cleanupPort(r.cfg.Port)
	}
}
