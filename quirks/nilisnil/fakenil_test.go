package nilisnil

import (
	"bufio"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHalfNilInterface(t *testing.T) {
	getNil := func() io.Writer {
		var writer *bufio.Writer // Pointer to a struct that implements io.Writer.

		assert.True(t, writer == nil, "Default value of an uninitialized pointer is nil.")
		return writer
	}

	n := getNil()
	assert.True(t, n == nil, "This is NOT nil, because the returned io.Writer interface has been initialized with the `*bufio.Writer` type even though it has a nil value.")
}
