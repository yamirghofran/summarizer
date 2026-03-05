package processor

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// videoExtensions contains supported video file extensions
var videoExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".m4v":  true,
}

// audioExtensions contains supported audio file extensions
var audioExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
	".flac": true,
	".ogg":  true,
	".aac":  true,
	".wma":  true,
}

// IsVideoFile checks if a file is a supported video format
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExtensions[ext]
}

// IsAudioFile checks if a file is a supported audio format
func IsAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return audioExtensions[ext]
}

// ConvertToAudio extracts audio from a video file and converts it to WAV format
// Returns the path to the converted WAV file
func ConvertToAudio(inputPath string) (string, error) {
	// Generate output filename (same directory, same basename, .wav extension)
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	outputPath := base + ".wav"

	// Build the ffmpeg command to extract audio
	cmd := exec.Command("ffmpeg",
		"-i", inputPath, // Input file
		"-ar", "44100", // Sample rate: 44.1kHz
		"-ac", "2", // Stereo audio
		"-f", "wav", // Output format: WAV
		"-y", // Overwrite output file if exists
		outputPath,
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed to extract audio: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}
