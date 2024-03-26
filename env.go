package kuu

import (
	"os"
	"strings"
)

func RuntimeEnv() string {
	return os.Getenv(ConfigEnv)
}

func IsProduction() bool {
	return strings.ToLower(RuntimeEnv()) == "production"
}
