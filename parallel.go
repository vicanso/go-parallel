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
	"reflect"
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
type RaceTask func() error

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
		wg.Add(1)
		ch <- struct{}{}
		go func(index int) {
			// 如果设置了出错时退出，而且当前出错数量不为1
			// 此处仅简单的使用atomic处理，并不保证当一个任务出错后，
			// 后续的任务肯定不执行。
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
		}(i)
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

// Race runs task function race, it's done when one task has been done
func Race(tasks ...RaceTask) error {
	size := len(tasks)
	if size == 0 {
		return nil
	}
	if size == 1 {
		return tasks[0]()
	}

	selectCaseList := make([]reflect.SelectCase, size)
	// 根据task生成对应的chan error
	for index, task := range tasks {
		ch := make(chan error)
		selectCaseList[index].Dir = reflect.SelectRecv
		selectCaseList[index].Chan = reflect.ValueOf(ch)
		go func(c chan error, t RaceTask) {
			c <- t()
		}(ch, task)
	}

	_, recv, recvOk := reflect.Select(selectCaseList)
	if !recvOk {
		return errors.New("receive from chan fail")
	}
	value := recv.Interface()
	if value == nil {
		return nil
	}
	err, ok := value.(error)
	if !ok {
		return errors.New("receive value should be error")
	}
	return err
}

// Some returns task function, when success time is >= count,
// it returns nil, otherwise returns error.
func Some(max, count int, fn Task) error {
	if max <= 0 || count <= 0 {
		return errors.New("max and count should be gt 0")
	}
	if count >= max {
		return errors.New("max should be gt count")
	}
	wg := sync.WaitGroup{}
	successCount := int32(0)
	done := make(chan int)
	errs := Errors{}
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func(index int) {
			err := fn(index)
			if err == nil {
				v := atomic.AddInt32(&successCount, 1)
				// 已达到完成条件
				if int(v) == count {
					done <- 0
				}
			} else {
				errs.Add(err)
			}
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		done <- 1
	}()

	// 成功的的任务数达到max
	if <-done == 0 {
		return nil
	}

	// 全部结束，但未达到成功要求
	return &errs
}

// Any runs task function, if one task is success, it will return nil,
// otherwise returns error.
func Any(max int, fn Task) error {
	return Some(max, 1, fn)
}
