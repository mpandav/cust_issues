package sendmail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
}

func TestRecipients(t *testing.T) {
	re := "123@tibco .com ,234@tibc.com"
	assert.Equal(t, []string{"123@tibco .com", "234@tibc.com"}, getRecipients(re))

	re2 := "123@tibco .com "
	assert.Equal(t, []string{"123@tibco .com"}, getRecipients(re2))
}
