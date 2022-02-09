package taskmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEngine_String will test the method String()
func TestEngine_String(t *testing.T) {

	t.Run("test all engine names", func(t *testing.T) {
		assert.Equal(t, "empty", Empty.String())
		assert.Equal(t, "machinery", Machinery.String())
		assert.Equal(t, "taskq", TaskQ.String())
	})
}

// TestEngine_IsEmpty will test the method IsEmpty()
func TestEngine_IsEmpty(t *testing.T) {
	t.Run("test empty engine", func(t *testing.T) {
		e := Empty
		assert.Equal(t, true, e.IsEmpty())
	})

	t.Run("test regular engine", func(t *testing.T) {
		e := Machinery
		assert.Equal(t, false, e.IsEmpty())
	})
}
