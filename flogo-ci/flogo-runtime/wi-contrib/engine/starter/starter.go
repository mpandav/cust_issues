package starter

import "fmt"

var shimRunner map[string]ShimStarter

const (
	TERMINATE          = 0
	WAIT_FOR_TERMINATE = 1
)

type ShimStarter interface {
	Init(args []string) error
	Run(args []string) (int, error)
}

func RegistryShimStarter(name string, m ShimStarter) error {
	if len(shimRunner) <= 0 {
		shimRunner = make(map[string]ShimStarter)
	}

	_, ok := shimRunner[name]
	if ok {
		return fmt.Errorf("main run [%s] already registered", name)
	}

	shimRunner[name] = m
	return nil
}

func HasShimRunner() bool {
	return len(shimRunner) > 0
}

func AllRunner() map[string]ShimStarter {
	return shimRunner
}
