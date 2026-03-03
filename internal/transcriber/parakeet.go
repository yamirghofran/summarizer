package transcriber

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Transcribe transcribes a WAV file using parakeet-mlx
// Returns the transcription text
func Transcribe(wavPath string) (string, error) {
	// Get the directory and filename
	inputDir := filepath.Dir(wavPath)
	inputBase := filepath.Base(wavPath)

	// Output file will be in the same directory with .txt extension
	transcriptPath := filepath.Join(inputDir, strings.TrimSuffix(inputBase, filepath.Ext(inputBase))+".txt")

	// Build the parakeet-mlx command
	// Use --output-dir to write to the same directory as input file
	cmd := exec.Command("parakeet-mlx",
		wavPath,
		"--model", "mlx-community/parakeet-tdt-0.6b-v3",
		"--output-format", "txt",
		"--output-dir", inputDir,
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("parakeet-mlx failed: %w\nOutput: %s", err, string(output))
	}

	// Read the transcription file
	content, err := os.ReadFile(transcriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcription file: %w", err)
	}

	// Clean up the transcription file
	defer os.Remove(transcriptPath)

	// Return the transcription text (trimmed)
	return strings.TrimSpace(string(content)), nil
}

// CheckBinary verifies that parakeet-mlx is available in PATH
func CheckBinary() error {
	_, err := exec.LookPath("parakeet-mlx")
	if err != nil {
		return fmt.Errorf("parakeet-mlx not found in PATH. Please install it first")
	}
	return nil
}
