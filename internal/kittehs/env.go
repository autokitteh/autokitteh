package kittehs

import "os"

func GetenvOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}
