package adminhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludeDomains(t *testing.T) {
	assert.NoError(t, validateHostnames([]string{"foo.crc.testing", "foo.bar.apps-crc.testing"}))
	assert.Error(t, validateHostnames([]string{"redhat.com"}))
}
