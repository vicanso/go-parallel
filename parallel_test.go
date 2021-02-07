package parallel

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParallel(t *testing.T) {
	assert := assert.New(t)

	arr := strings.Split("0123456789", "")
	count := int32(0)
	sum := int32(0)
	fn := func(index int) error {
		atomic.AddInt32(&sum, int32(index))
		atomic.AddInt32(&count, 1)
		return errors.New(strconv.Itoa(index))
	}

	err := Parallel(len(arr), 3, fn)
	assert.NotNil(err)
	errs, ok := err.(*Errors)
	assert.True(ok)
	assert.Equal(len(arr), len(errs.Errs))
	assert.Equal(len(arr), int(count))
	assert.Equal(45, int(sum))
}
