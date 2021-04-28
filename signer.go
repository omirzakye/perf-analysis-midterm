package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var SingleHash = func(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go func(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			data := strconv.Itoa(in.(int))
			mu.Lock()
			datamd5 := DataSignerMd5(data)
			mu.Unlock()
			ch := make(chan string)
			go crc32Func(data, ch)
			crc32Md5res := DataSignerCrc32(datamd5)
			crc32res := <- ch
			fmt.Printf("SingleHash data %s\n", data)
			fmt.Printf("SingleHash crc32(data) %s\n", crc32res)
			fmt.Printf("SingleHash md5(data) %s\n", datamd5)
			fmt.Printf("SingleHash crc32(md5(data)) %s\n", crc32Md5res)
			fmt.Printf("SingleHash result %s\n", crc32res+"~"+crc32Md5res)
			out <- crc32res + "~" + crc32Md5res
		}(i, out, wg, mu)
	}
	wg.Wait()
}

func crc32Func(data string, out chan<- string) {
	out <- DataSignerCrc32(data)
}

var MultiHash = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)

		go func(s string, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			mu := &sync.Mutex{}
			wg2 := &sync.WaitGroup{}
			resArray := make([]string, 6)

			for i := 0; i < 6; i++ {
				wg2.Add(1)
				data := strconv.Itoa(i) + s
				go mh(s, resArray, data, i, wg2, mu)
			}

			wg2.Wait()
			result := strings.Join(resArray, "")
			fmt.Printf("MultiHash result: %s\n", result)
			out <- result
		}(i.(string), out, wg)
	}
	wg.Wait()
}

func mh(s string, resArray []string, data string, i int, wg *sync.WaitGroup, mu *sync.Mutex){
	defer wg.Done()
	data = DataSignerCrc32(data)
	mu.Lock()
	resArray[i] = data
	fmt.Printf("MultiHash: crc32(th+step1)) %d %s\n", i, data)
	mu.Unlock()
}

var CombineResults = func(in, out chan interface{}) {
	var mystrings []string
	for i := range in {
		mystrings = append(mystrings, i.(string))
	}
	sort.Strings(mystrings)
	res := strings.Join(mystrings, "_")
	fmt.Printf("CombinedResults \n%s\n", res)
	out <- res
}

var ExecutePipeline = func(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, j := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go jobWorker(j, in, out, wg)
		in = out
	}
	wg.Wait()
}

func jobWorker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	job(in, out)
}