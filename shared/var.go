package shared

import (
	"log"
	"os"
	"path/filepath"

	"github.com/zhquiz/go-zhquiz/server/rand"
)

// UserDataDir is used to store all writable data
func UserDataDir() string {
	return os.Getenv("USER_DATA_DIR")
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

// APISecret returns ZHQUIZ_API_SECRET for programmatic API access
func APISecret() string {
	return getenvOrDefaultFn("ZHQUIZ_API_SECRET", func() string {
		s, err := rand.GenerateRandomString(64)
		if err != nil {
			log.Fatalln(err)
		}
		return s
	})
}

// DatabaseURL returns DATABASE_URL
func DatabaseURL() string {
	return getenvOrDefaultFn("DATABASE_URL", func() string {
		return filepath.Join(UserDataDir(), "data.db")
	})
}

// SpeakFn return ZHQUIZ_SPEAK for programmatic speak function
func SpeakFn() string {
	s := getenvOrDefaultFn("ZHQUIZ_SPEAK", func() string {
		defaultPath := filepath.Join(UserDataDir(), "speak.sh")

		stat, err := os.Stat(defaultPath)
		if err == nil && !stat.IsDir() {
			return defaultPath
		}

		return "0"
	})

	if s == "0" {
		return ""
	}

	return s
}
