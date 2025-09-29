package builder

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/patnaikankit/Gaze/internal/config"
	"github.com/patnaikankit/Gaze/pkg/logger"
)

type Builder struct {
	cfg *config.Config
}

type BuildResult struct {
	Success  bool
	Duration time.Duration
	Output   string
	BuildErr error
}

func New(cfg *config.Config) *Builder {
	return &Builder{cfg: cfg}
}

func (b *Builder) Build() BuildResult {
	start := time.Now()

	// Ensure output directory exists
	if strings.Contains(b.cfg.BuildCMD, string(filepath.Separator)+"temp"+string(filepath.Separator)) {
		if err := os.MkdirAll("./temp", 0o755); err != nil {
			logger.Error("Failed to create build temp directory: %v", err)
		}
	}

	// Split command into executable + args
	parts := strings.Fields(b.cfg.BuildCMD)
	if len(parts) == 0 {
		logger.Error("No build command provided")
		return BuildResult{Success: false, BuildErr: nil, Output: "invalid build command"}
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		logger.Error("Build failed in %v", duration)
		return BuildResult{
			Success:  false,
			Duration: duration,
			Output:   string(output),
			BuildErr: err,
		}
	}

	logger.Info("Build succeeded in %v", duration)
	return BuildResult{
		Success:  true,
		Duration: duration,
		Output:   string(output),
	}
}
