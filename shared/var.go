package shared

import (
	"log"
	"os"
	"path/filepath"
)

// UserDataDir is used to store all writable data
func UserDataDir() string {
	dir := os.Getenv("USER_DATA_DIR")
	if dir == "" {
		return ExecDir
	}

	return dir
}

// MediaPath returns path to media folder, and mkdir if necessary
func MediaPath() string {
	mediaPath := filepath.Join(UserDataDir(), "_media")
	_, err := os.Stat(mediaPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(mediaPath, 0644); err != nil {
			log.Fatalln(err)
		}
	}

	return mediaPath
}

// Port gets server port, or "default" port value
func Port() string {
	return getenvOrDefault("PORT", "35594")
}

// DatabaseURL returns DATABASE_URL
func DatabaseURL() string {
	return getenvOrDefaultFn("DATABASE_URL", func() string {
		return filepath.Join(UserDataDir(), "data.db")
	})
}

// IsDebug decides whether to run in debug mode (e.g. development server)
func IsDebug() bool {
	return os.Getenv("DEBUG") != ""
}
