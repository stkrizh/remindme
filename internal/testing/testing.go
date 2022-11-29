package testing

import "os"

func IsIntegration() bool {
	return os.Getenv("TEST_INTEGRATION") == "1"
}
