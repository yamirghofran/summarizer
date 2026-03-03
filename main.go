package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/yamirghofran/summarizer/cmd"
)

func main() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Also try loading from current directory
	if _, err := os.Stat(".env"); err != nil {
		// Try loading from home directory as fallback
		homeDir, err := os.UserHomeDir()
		if err == nil {
			_ = godotenv.Load(homeDir + "/.summarizer.env")
		}
	}

	cmd.Execute()
}
