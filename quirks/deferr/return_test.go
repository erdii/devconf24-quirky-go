package deferr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expected = errors.New("expected")
var defered = errors.New("defered")
var unexpected = errors.New("spanish inquisition")

func TestReturn(t *testing.T) {
	actual := func() error {
		var err error
		defer func() {
			err = expected
		}()
		return err
	}()

	assert.Equal(t, expected, actual)
}

func TestNamedReturn(t *testing.T) {
	actual := func() (err error) {
		defer func() {
			err = expected
		}()
		return err
	}()

	assert.Equal(t, expected, actual)
}

func TestNamedReturnShadow(t *testing.T) {
	actual := func() (err error) {
		err = expected
		defer func() {
			err = unexpected
		}()
		return err
	}()

	assert.Equal(t, expected, actual)
}

func TestNamedReturnWrap(t *testing.T) {
	actual := func() (err error) {
		err = expected
		defer func() {
			cErr := defered
			if err != nil && cErr != nil {
				err = fmt.Errorf("defered func errored, but surrounding function also encountered an error.\nsurrounding err: %w, defered err: %w", err, cErr)
			} else if cErr != nil {
				err = cErr
			}
		}()
		return err
	}()

	assert.ErrorIs(t, actual, expected)
	assert.ErrorIs(t, actual, defered)
}
