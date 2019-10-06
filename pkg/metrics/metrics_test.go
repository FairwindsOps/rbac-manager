package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterMetrics(t *testing.T) {
	assert.NotPanics(t, RegisterMetrics)
}
