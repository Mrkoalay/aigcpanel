package utils

import "os"

func GetEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}
func Contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}
