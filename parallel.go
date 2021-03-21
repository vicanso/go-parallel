// Copyright 2021 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parallel

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
)

type Option struct {
	// the size of tasks
	Max int
	// the parallel limit
	Limit int
	// task function
	Task Task
	// break the function on error
	BreakOnError bool
}
type Task func(index int) error

// Errors
type Errors struct {
	mutex sync.Mutex
	Errs  []error
}

// Add add error
func (errs *Errors) Add(err error) {
	errs.mutex.Lock()
	defer errs.mutex.Unlock()
	errs.Errs = append(errs.Errs, err)
}

// Exists check error is exists
func (errs *Errors) Exists() bool {
	return len(errs.Errs) != 0
}

func (errs *Errors) Error() string {
	arr := make([]string, len(errs.Errs))
	for index, err := range errs.Errs {
		arr[index] = err.Error()
	}
	return strings.Join(arr, ", ")
}

// EnhancedParallel runs the task function parallel
func EnhancedParallel(opt Option) error {
	if opt.Limit <= 0 || opt.Max <= 0 {
		return errors.New("max and limit should be gt 0")
	}
	if opt.Task == nil {
		return errors.New("task function can not be nil")
	}
	errs := &Errors{}
	errCount := int32(0)

	// 设置带缓存的channel
	ch := make(chan struct{}, opt.Limit)
	var wg sync.WaitGroup
	for i := 0; i < opt.Max; i++ {
		index := i
		wg.Add(1)
		ch <- struct{}{}
		go func() {
			// 如果设置了出错时退出，而且当前出错数量不为1
			if opt.BreakOnError && atomic.LoadInt32(&errCount) != 0 {
				wg.Done()
				<-ch
				return
			}
			defer func() {
				wg.Done()
				<-ch
			}()
			err := opt.Task(index)
			if err != nil {
				if opt.BreakOnError {
					atomic.AddInt32(&errCount, 1)
				}
				errs.Add(err)
			}
		}()
	}
	// 等待所有任务完成
	wg.Wait()
	// 关闭channel
	close(ch)
	if errs.Exists() {
		return errs
	}
	return nil
}

// Parallel runs task function parallel
func Parallel(max, limit int, fn Task) error {
	return EnhancedParallel(Option{
		Max:   max,
		Limit: limit,
		Task:  fn,
	})
}
