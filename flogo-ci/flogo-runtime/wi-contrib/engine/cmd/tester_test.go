package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFlowName(t *testing.T) {
	url := "res://flow:SendMessage"
	name := getFlowName(url)
	assert.Equal(t, "SendMessage", name)
}
