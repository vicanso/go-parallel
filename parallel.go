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
)

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

// Parallel limit task running for parallel call
func Parallel(max, limit int, fn Task) error {
	if limit <= 0 || max <= 0 {
		return errors.New("max and limit should be gt 0")
	}

	errs := &Errors{}

	// 设置带缓存的channel
	ch := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i := 0; i < max; i++ {
		index := i
		wg.Add(1)
		ch <- struct{}{}
		go func() {
			defer func() {
				wg.Done()
				<-ch
			}()
			err := fn(index)
			if err != nil {
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
