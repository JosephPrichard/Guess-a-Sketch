package utils

import (
	"fmt"
	"os"
)

func GetEnvVar(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("Environment variable %s wasn't set, have you checked your .env file?", k))
	}
	return v
}