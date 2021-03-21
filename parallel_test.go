package parallel

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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
		return nil
	}

	err := Parallel(len(arr), 3, fn)
	assert.Nil(err)
	assert.Equal(len(arr), int(count))
	assert.Equal(45, int(sum))
}

func TestParallelError(t *testing.T) {
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

func TestParallelBreakOnError(t *testing.T) {

	assert := assert.New(t)

	arr := strings.Split("0123456789", "")
	count := int32(0)
	sum := int32(0)
	fn := func(index int) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&sum, int32(index))
		atomic.AddInt32(&count, 1)
		return errors.New(strconv.Itoa(index))
	}

	limit := 5
	err := EnhancedParallel(Option{
		Max:          len(arr),
		Limit:        limit,
		Task:         fn,
		BreakOnError: true,
	})
	// 由于设置了出错则退出，因此只执行了一次limit数量的task任务
	assert.NotNil(err)
	errs, ok := err.(*Errors)
	assert.True(ok)
	assert.Equal(limit, len(errs.Errs))
	assert.Equal(limit, int(count))
	assert.Equal(10, int(sum))
}

func TestErrors(t *testing.T) {
	assert := assert.New(t)

	errs := Errors{}
	errs.Add(errors.New("abc"))

	assert.True(errs.Exists())
	assert.Equal("abc", errs.Error())
}
