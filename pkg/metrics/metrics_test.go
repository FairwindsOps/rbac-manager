package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterMetrics(t *testing.T) {
	assert.NotPanics(t, RegisterMetrics)
}
