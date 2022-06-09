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
	// lock task function
	LockTask LockTask
	// break the function on error
	BreakOnError bool
}
type Task func(index int) error
type LockTask func(index int, rw *sync.RWMutex) error
type RaceTask func() error

// Errors
type Errors struct {
	mutex sync.Mutex
	Errs  []error
}

// Add add error
func (errs *Errors) Add(err error) {
	if err == nil {
		return
	}
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
	if opt.Task == nil && opt.LockTask == nil {
		return errors.New("task(lock task) function can not be nil")
	}
	errs := &Errors{}
	errCount := int32(0)

	// 设置带长度的channel，用于限制并发调用
	ch := make(chan struct{}, opt.Limit)
	var wg sync.WaitGroup

	rwMutex := &sync.RWMutex{}

	shouldBreak := func() bool {
		return opt.BreakOnError && atomic.LoadInt32(&errCount) != 0
	}

	for i := 0; i < opt.Max; i++ {
		// 如果设置出错时退出，而且出错数量不为0
		// 直接退出当前循环
		if shouldBreak() {
			break
		}
		wg.Add(1)
		// 此处会限制最多只创建limit数量的goroutine
		ch <- struct{}{}
		go func(index int) {
			defer func() {
				wg.Done()
				<-ch
			}()
			// 如果设置了出错时退出，而且当前出错数量不为1
			// 此处仅简单的使用atomic处理，并不保证当一个任务出错后，
			// 后续的任务肯定不执行。
			if shouldBreak() {
				return
			}
			var err error
			if opt.LockTask != nil {
				err = opt.LockTask(index, rwMutex)
			} else {
				err = opt.Task(index)
			}

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
	// 关闭channel（不关闭则在回收时自动清除）
	close(ch)
	if errs.Exists() {
		return errs
	}
	return nil
}

func enhancedParallel(fn Task, lockFn LockTask, max int, limit ...int) error {
	l := 2
	if len(limit) != 0 && limit[0] > 0 {
		l = limit[0]
	}
	return EnhancedParallel(Option{
		Max:          max,
		Limit:        l,
		Task:         fn,
		LockTask:     lockFn,
		BreakOnError: true,
	})
}

// Parallel runs task function parallel
func Parallel(fn Task, max int, limit ...int) error {
	return enhancedParallel(fn, nil, max, limit...)
}

// ParallelLock runs lock task function parallel
func ParallelLock(fn LockTask, max int, limit ...int) error {
	return enhancedParallel(nil, fn, max, limit...)
}

func race(onlySucceed bool, tasks ...RaceTask) error {
	size := len(tasks)
	if size == 0 {
		return nil
	}
	if size == 1 {
		return tasks[0]()
	}

	doneCount := int32(0)
	successCount := int32(0)
	var err error
	done := make(chan struct{}, 1)
	defer close(done)
	taskCount := len(tasks)
	// 根据task生成对应的chan error
	for _, task := range tasks {
		go func(t RaceTask) {
			e := t()
			value := atomic.AddInt32(&doneCount, 1)
			// 如果非判断成功的任务
			if !onlySucceed {
				// 只有第一个完成的才会触发
				if value == 1 {
					err = e
					done <- struct{}{}
				}
				return
			}

			// 成功的第一个任务
			if e == nil && atomic.AddInt32(&successCount, 1) == 1 {
				done <- struct{}{}
				return
			}

			// 所有任务均完成，但无一成功
			if value == int32(taskCount) && atomic.LoadInt32(&successCount) == 0 {
				err = e
				done <- struct{}{}
			}
		}(task)
	}
	<-done
	return err
}

// Race runs task function race, it's done when one task has been done
func Race(tasks ...RaceTask) error {
	return race(false, tasks...)
}

// RaceSucceed runs task function race, it's done when one task has been successful or all task done.
func RaceSucceed(tasks ...RaceTask) error {
	return race(true, tasks...)
}

// Some returns task function, when success time is >= count,
// it returns nil, otherwise returns error.
func Some(fn Task, max, count int) error {
	if max <= 0 || count <= 0 {
		return errors.New("max and count should be gt 0")
	}
	if count >= max {
		return errors.New("max should be gt count")
	}
	successCount := int32(0)
	doneCount := int32(0)
	done := make(chan int, 1)
	defer close(done)
	errs := &Errors{}
	for i := 0; i < max; i++ {
		go func(index int) {
			err := fn(index)
			if err == nil {
				v := atomic.AddInt32(&successCount, 1)
				// 已达到完成条件，设置完成，并直接返回
				// 不再触发doneCount，避免successCount与doneCount同时满足
				if int(v) == count {
					done <- 0
					return
				}
			}
			errs.Add(err)
			// 已执行完最后一条，但未满足successCount
			if int(atomic.AddInt32(&doneCount, 1)) == max {
				done <- 1
			}
		}(i)
	}

	// 成功的的任务数达到max
	if <-done == 0 {
		return nil
	}

	// 全部结束，但未达到成功要求
	return errs
}

// Any runs task function, if one task is success, it will return nil,
// otherwise returns error.
func Any(fn Task, max int) error {
	return Some(fn, max, 1)
}
