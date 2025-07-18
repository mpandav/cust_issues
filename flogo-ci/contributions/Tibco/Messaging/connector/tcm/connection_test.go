package tcm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRetryLogic(t *testing.T) {
	re := &retry{
		attempts:          0,
		maxDelay:          5 * time.Second,
		autoReconAttempts: 6,
	}

	var err error
	err = re.retry(err, call)
	assert.NotNil(t, err)
}

func call() error {
	return fmt.Errorf("Error to trigger retry")
}

func TestRetryLogic2(t *testing.T) {
	re := &retry{
		attempts:          0,
		maxDelay:          5 * time.Second,
		autoReconAttempts: 6,
	}

	var err error
	err = re.retry(err, call2)
	assert.Nil(t, err)
}

func call2() error {
	return nil
}
