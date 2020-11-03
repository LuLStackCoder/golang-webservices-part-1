package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SingleHash ...
func SingleHash(in, out chan interface{}) {
	start := time.Now()
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for i := range in {
		md5Crc32Hash := make(chan string)
		crc32Hash := make(chan string)
		value, ok := i.(string)
		if !ok {
			value = strconv.Itoa(i.(int))
		}
		wg.Add(1)
		go func(val string) {
			defer wg.Done()
			go func() {
				mu.Lock()
				md5Hash := DataSignerMd5(val)
				mu.Unlock()
				md5Crc32Hash <- DataSignerCrc32(md5Hash)
			}()
			go func() {
				crc32Hash <- DataSignerCrc32(val)
			}()
			res := <-crc32Hash + "~" + <-md5Crc32Hash
			fmt.Println(res)
			out <- res
		}(value)
	}
	wg.Wait()
	end := time.Since(start)
	fmt.Println(end)
}

// MultiHash ...
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		value, ok := i.(string)
		if !ok {
			value = strconv.Itoa(i.(int))
		}
		wg.Add(1)
		go func(val string) {
			defer wg.Done()
			wgWorker := &sync.WaitGroup{}
			hashSlice := make([]string, 6)
			for i := 0; i < 6; i++ {
				wgWorker.Add(1)
				go func(th int) {
					defer wgWorker.Done()
					hashSlice[th] = DataSignerCrc32(strconv.Itoa(th) + val)
				}(i)
			}
			wgWorker.Wait()
			res := strings.Join(hashSlice, "")
			out <- res
		}(value)
	}
	wg.Wait()
}

// CombineResults ...
func CombineResults(in, out chan interface{}) {
	hashSlice := make([]string, 0)
	for i := range in {
		hashSlice = append(hashSlice, i.(string))
	}
	sort.Strings(hashSlice)
	res := strings.Join(hashSlice, "_")
	out <- res
}

// ExecutePipeline ...
func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 100)
	wg := &sync.WaitGroup{}
	for _, j := range jobs {
		wg.Add(1)
		out := make(chan interface{}, 100)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(j, in, out)
		in = out
	}

	wg.Wait()
}
