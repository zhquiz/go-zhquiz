package shared

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
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
func Port() int {
	defaultPort := 35594

	port := getenvOrSetDefault("PORT", strconv.Itoa(defaultPort))
	if p, e := strconv.Atoi(port); e == nil {
		return p
	}

	return defaultPort
}

// IsDebug decides whether to run in debug mode (e.g. development server)
func IsDebug() bool {
	return os.Getenv("DEBUG") != ""
}

// IsChromeApp decides whether to run in Chrome App (i.e. windowed mode)
func IsChromeApp() bool {
	return getenvOrSetDefault("ZHQUIZ_CHROME_APP", "1") != "0"
}
