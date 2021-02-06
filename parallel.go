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
	"sync"
)

type Task func(index int)

// Parallel limit task running for parallel call
func Parallel(max, limit int, fn Task) error {
	if limit <= 0 || max <= 0 {
		return errors.New("max and limit should be gt 0")
	}

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
			fn(index)
		}()
	}
	// 等待所有任务完成
	wg.Wait()
	// 关闭channel
	close(ch)
	return nil
}
