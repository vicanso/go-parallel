package parallel

import (
	"errors"
	"strconv"
	"strings"
	"sync"
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

	err := Parallel(fn, len(arr))
	assert.Nil(err)
	assert.Equal(len(arr), int(count))
	assert.Equal(45, int(sum))
}

func TestParallelLock(t *testing.T) {
	assert := assert.New(t)

	arr := strings.Split("0123456789", "")
	count := 0
	sum := 0
	fn := func(index int, rw *sync.RWMutex) error {
		rw.Lock()
		defer rw.Unlock()
		count++
		sum += index
		return nil
	}

	err := ParallelLock(fn, len(arr))
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

	err := Parallel(fn, len(arr), 3)
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

func TestRace(t *testing.T) {
	assert := assert.New(t)

	err := Race()
	assert.Nil(err)
	customErr := errors.New("custom error")

	err = Race(func() error {
		return customErr
	})
	assert.Equal(customErr, err)

	err = Race(func() error {
		return customErr
	}, func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	assert.Equal(customErr, err)

	err = Race(func() error {
		time.Sleep(50 * time.Millisecond)
		return customErr
	}, func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	assert.Nil(err)
}

func TestSome(t *testing.T) {
	assert := assert.New(t)

	err := Some(func(index int) error {
		return errors.New("error")
	}, 5, 3)
	assert.NotNil(err)
	errs, ok := err.(*Errors)
	assert.True(ok)
	assert.Equal(5, len(errs.Errs))

	err = Some(func(index int) error {
		if index%2 == 0 {
			return nil
		}
		return errors.New("error")
	}, 5, 3)
	assert.Nil(err)
}

func TestAny(t *testing.T) {
	assert := assert.New(t)

	err := Any(func(index int) error {
		return errors.New("error")
	}, 5)
	assert.NotNil(err)
	errs, ok := err.(*Errors)
	assert.True(ok)
	assert.Equal(5, len(errs.Errs))

	err = Any(func(index int) error {
		if index == 4 {
			return nil
		}
		return errors.New("error")
	}, 5)
	assert.Nil(err)
}

func TestErrors(t *testing.T) {
	assert := assert.New(t)

	errs := Errors{}
	errs.Add(errors.New("abc"))

	assert.True(errs.Exists())
	assert.Equal("abc", errs.Error())
}
