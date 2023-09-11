package utils

import (
	"fmt"
	"log"
	"os"
)

func GetEnvVar(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("Environment variable %s wasn't set", k))
	}
	log.Printf("Environment variable %s %s", k, v)
	return v
}