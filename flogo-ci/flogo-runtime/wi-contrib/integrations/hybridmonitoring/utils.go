package hybridmonitoring

import (
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func getCPUFromLibrary() float32 {
	cpuinfo, _ := cpu.Percent(time.Second, false)
	return float32(cpuinfo[0])
}

func extractCPU(top string) (float32, error) {
	cpus := strings.Fields(top)
	user, err := strconv.ParseFloat(strings.Trim(cpus[0], "%"), 32)
	sys, err := strconv.ParseFloat(strings.Trim(cpus[1], "%"), 32)

	return float32(user + sys), err

}
