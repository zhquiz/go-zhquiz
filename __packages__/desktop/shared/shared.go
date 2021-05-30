package shared

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// ExecDir is dirname of executable
var ExecDir string

func init() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	ExecDir = dir
}

// Load loads from .env and .env.local to os.Getenv
func Load() {
	godotenv.Load(filepath.Join(ExecDir, ".env"), filepath.Join(ExecDir, ".env.local"))
}

// Setenv sets to .env.local and os.Getenv
func setenv(key string, value string) {
	env, _ := godotenv.Read(filepath.Join(ExecDir, ".env.local"))
	env[key] = value
	godotenv.Write(env, filepath.Join(ExecDir, ".env.local"))

	os.Setenv(key, value)
}

// getenvOrSetDefault writes to .env if env not exists
func getenvOrSetDefault(key string, value string) string {
	v := os.Getenv(key)
	if v == "" {
		v = value
		setenv(key, v)
	}

	return v
}

// getenvOrSetDefaultFn writes to .env if env not exists, using function
func getenvOrSetDefaultFn(key string, fn func() string) string {
	v := os.Getenv(key)
	if v == "" {
		v = fn()
		setenv(key, v)
	}

	return v
}
