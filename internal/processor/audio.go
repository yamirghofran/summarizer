package processor

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CompressAndSpeedUp compresses the audio file and speeds it up using ffmpeg
// Converts to 16kHz mono, 32kbps, and applies 1.7x tempo
func CompressAndSpeedUp(inputPath string) (string, error) {
	// Generate output filename
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	outputPath := base + "-fast-compressed" + ext

	// Build the ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ar", "16000", // Sample rate: 16kHz
		"-ac", "1", // Mono audio
		"-ab", "32k", // Bitrate: 32kbps
		"-f", "wav", // Output format: WAV
		"-filter:a", "atempo=1.7", // Speed up 1.7x
		"-y", // Overwrite output file if exists
		outputPath,
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}

// CheckBinary verifies that ffmpeg is available in PATH
func CheckBinary() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH. Please install it first: https://ffmpeg.org")
	}
	return nil
}
