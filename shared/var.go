package shared

import (
	"log"
	"os"
	"path/filepath"

	"github.com/zhquiz/go-zhquiz/server/rand"
)

// MediaPath returns path to media folder, and mkdir if necessary
func MediaPath() string {
	mediaPath := filepath.Join(ExecDir, "_media")
	_, err := os.Stat(mediaPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(mediaPath, 0644); err != nil {
			log.Fatal(err)
		}
	}

	return mediaPath
}

// Port gets server port, or "default" port value
func Port() string {
	return getenvOrDefault("PORT", "35594")
}

// IsDesktop decides whether to run in desktop mode
func IsDesktop() bool {
	return os.Getenv("ZHQUIZ_DESKTOP") != "0"
}

// DatabaseURL returns DATABASE_URL
func DatabaseURL() string {
	return getenvOrDefaultFn("DATABASE_URL", func() string {
		paths := []string{"data.db"}
		if root := ExecDir; root != "" {
			paths = append([]string{root}, paths...)
		}

		return filepath.Join(paths...)
	})
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

// SpeakFn return ZHQUIZ_SPEAK for programmatic speak function
func SpeakFn() string {
	s := getenvOrDefaultFn("ZHQUIZ_SPEAK", func() string {
		defaultPath := filepath.Join(ExecDir, "speak.sh")

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

// CotterAPIKey returns COTTER_API_KEY for logging in
func CotterAPIKey() string {
	return os.Getenv("COTTER_API_KEY")
}

// DefaultUser returns DEFAULT_USER for use in Desktop mode
func DefaultUser() string {
	return os.Getenv("DEFAULT_USER")
}

// Plausible returns domain name for Plausible Analytics
func Plausible() string {
	return os.Getenv("PLAUSIBLE")
}
