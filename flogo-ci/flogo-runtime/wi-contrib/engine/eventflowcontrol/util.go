package eventflowcontrol

import (
	"os"
	"runtime"
	"strconv"

	"github.com/project-flogo/core/support/log"
)

func getFloatValueFromEnvVariable(name string, defaultVal float64, logger log.Logger) float64 {
	value := os.Getenv(name)
	if value != "" {
		ht, err := strconv.Atoi(value)
		if err != nil {
			logger.Warnf("Invalid value [%s] set for [%s]. It must be a valid percentage value like 90, 90.5 etc. Setting it to default [%0.2f%%].", value, name, defaultVal)
		} else {
			return float64(ht)
		}
	}
	return defaultVal
}

func getIntValueFromEnvVariable(name string, defaultVal int, logger log.Logger) int {
	value := os.Getenv(name)
	if value != "" {
		ht, err := strconv.Atoi(value)
		if err != nil {
			logger.Warnf("Invalid value [%s] set for [%s]. It must be a valid integer value. Setting it to default [%d].", value, name, defaultVal)
		} else {
			return ht
		}
	}
	return defaultVal
}

func isLinuxContainer() bool {
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
			return false
		}
		return true
	}
	return false
}
