/*
Copyright 2020 Deepak S<deepaks@outlook.in>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package goworkers

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFunctionalityWithoutArgs(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	gw.Stop()
}

func TestFunctionalityCheckErrorWithoutArgs(t *testing.T) {
	edone := make(chan struct{})
	errResps := 0
	errVals := 10

	gw := New()

	go func() {
		m := make(map[string]int, errVals)
		for err := range gw.ErrChan {
			s := err.(error).Error()
			m[s]++
			if m[s] > 1 {
				t.Errorf("Inconsistent value in error channel. Got %s more than once", s)
			}
			errResps++
		}
		edone <- struct{}{}
	}()

	fn := func(i int) error {
		return fmt.Errorf("e%d", i)
	}

	for val := 0; val < errVals; val++ {
		i := val
		gw.SubmitCheckError(func() error {
			return fn(i)
		})
	}

	gw.Stop()

	<-edone

	if errResps != errVals {
		t.Errorf("Expected %d error responses, got %d", errVals, errResps)
	}
}

func TestFunctionalityCheckResultWithoutArgs(t *testing.T) {
	edone := make(chan struct{})
	rdone := make(chan struct{})
	var (
		errResps int32
		errVals  int32
		resResps int32
		resVals  int32
	)
	rNum := 20

	gw := New()

	go func() {
		m := make(map[error]int, rNum)
		for err := range gw.ErrChan {
			m[err.(error)]++
			if m[err.(error)] > 1 {
				t.Errorf("Inconsistent value in error channel. Got %s more than once", err.Error())
			}
			errResps++
		}
		edone <- struct{}{}
	}()

	go func() {
		m := make(map[string]int, rNum)
		for val := range gw.ResultChan {
			m[val.(string)]++
			if m[val.(string)] > 1 {
				t.Errorf("Inconsistent value in result channel. Got %s more than once", val)
			}
			resResps++
		}
		rdone <- struct{}{}
	}()

	fn := func(i int) (interface{}, error) {
		if i%2 == 0 {
			atomic.AddInt32(&errVals, 1)
			return nil, fmt.Errorf("e%d", i)
		}
		atomic.AddInt32(&resVals, 1)
		return fmt.Sprintf("v%d", i), nil
	}

	for val := 0; val < rNum; val++ {
		i := val
		gw.SubmitCheckResult(func() (interface{}, error) {
			return fn(i)
		})
	}

	gw.Stop()

	<-edone
	<-rdone

	if errResps != errVals {
		t.Errorf("Expected %d error responses, got %d", errVals, errResps)
	}

	if resResps != resVals {
		t.Errorf("Expected %d result responses, got %d", resVals, resResps)
	}
}

func TestFunctionalityCheckMultiInstances(t *testing.T) {
	// gw1
	edonegw1 := make(chan struct{}, 1)
	rdonegw1 := make(chan struct{}, 1)
	var (
		errRespsgw1 int32
		errValsgw1  int32
		resRespsgw1 int32
		resValsgw1  int32
	)
	// gw2
	edonegw2 := make(chan struct{}, 1)
	rdonegw2 := make(chan struct{}, 1)
	var (
		errRespsgw2 int32
		errValsgw2  int32
		resRespsgw2 int32
		resValsgw2  int32
	)

	rNum := 500

	gw1 := New()
	gw2 := New()

	// gw1 output
	go func() {
		errList := []error{}
		for err := range gw1.ErrChan {
			errRespsgw1++
			errList = append(errList, err.(error))
		}

		m := make(map[string]int, rNum)
		for _, err := range errList {
			serr := err.Error()
			if !strings.HasPrefix(serr, "gw1") {
				t.Errorf("Received %s from worker gw2, expected values only from gw1", serr)
			}
			m[serr]++
			if m[serr] > 1 {
				t.Errorf("Inconsistent value in error channel. Got %s more than once", serr)
			}
		}
		edonegw1 <- struct{}{}
	}()

	go func() {
		m := make(map[string]int, rNum)
		for val := range gw1.ResultChan {
			sval := val.(string)
			if !strings.HasPrefix(sval, "gw1") {
				t.Errorf("Received %s from worker gw2, expected values only from gw1", sval)
			}
			m[sval]++
			if m[sval] > 1 {
				t.Errorf("Inconsistent value in result channel. Got %s more than once", val)
			}
			resRespsgw1++
		}
		rdonegw1 <- struct{}{}
	}()

	// gw2 output
	go func() {
		m := make(map[string]int, rNum)
		for err := range gw2.ErrChan {
			serr := err.(error).Error()
			if !strings.HasPrefix(serr, "gw2") {
				t.Errorf("Received %s from worker gw1, expected values only from gw2", serr)
			}
			m[serr]++
			if m[serr] > 1 {
				t.Errorf("Inconsistent value in error channel. Got %s more than once", serr)
			}
			errRespsgw2++
		}
		edonegw2 <- struct{}{}
	}()

	go func() {
		m := make(map[string]int, rNum)
		for val := range gw2.ResultChan {
			sval := val.(string)
			if !strings.HasPrefix(sval, "gw2") {
				t.Errorf("Received %s from worker gw1, expected values only from gw2", sval)
			}
			m[sval]++
			if m[sval] > 1 {
				t.Errorf("Inconsistent value in result channel. Got %s more than once", val)
			}
			resRespsgw2++
		}
		rdonegw2 <- struct{}{}
	}()

	// gw1 func()
	fngw1 := func(i int) (interface{}, error) {
		if i%2 == 0 {
			atomic.AddInt32(&errValsgw1, 1)
			return nil, fmt.Errorf("gw1e%d", i)
		}
		atomic.AddInt32(&resValsgw1, 1)
		return fmt.Sprintf("gw1v%d", i), nil
	}

	// gw2 func()
	fngw2 := func(i int) (interface{}, error) {
		if i%2 == 0 {
			atomic.AddInt32(&errValsgw2, 1)
			return nil, fmt.Errorf("gw2e%d", i)
		}
		atomic.AddInt32(&resValsgw2, 1)
		return fmt.Sprintf("gw2v%d", i), nil
	}

	// gw1 submission
	for val := 0; val < rNum; val++ {
		i := val
		gw1.SubmitCheckResult(func() (interface{}, error) {
			return fngw1(i)
		})
	}

	// gw2 submission
	for val := 0; val < rNum; val++ {
		i := val
		gw2.SubmitCheckResult(func() (interface{}, error) {
			return fngw2(i)
		})
	}

	gw1.Stop()
	gw2.Stop()

	<-edonegw1
	<-rdonegw1

	<-edonegw2
	<-rdonegw2

	if errRespsgw1 != errValsgw1 {
		t.Errorf("Expected %d error responses, got %d", errValsgw1, errRespsgw1)
	}

	if resRespsgw1 != resValsgw1 {
		t.Errorf("Expected %d result responses, got %d", resValsgw1, resRespsgw1)
	}

	if errRespsgw2 != errValsgw2 {
		t.Errorf("Expected %d error responses, got %d", errValsgw2, errRespsgw2)
	}

	if resRespsgw2 != resValsgw2 {
		t.Errorf("Expected %d result responses, got %d", resValsgw2, resRespsgw2)
	}
}

func TestFunctionalityWithArgs(t *testing.T) {
	opts := Options{Workers: 3}
	gw := New(opts)

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	gw.Stop()
}

func TestWorkerArg(t *testing.T) {
	tables := []struct {
		Given    uint32
		Expected uint32
	}{
		{1, 1},
		{2, 2},
		{0, 0},
	}

	for _, table := range tables {
		opts := Options{Workers: table.Given}
		gw := New(opts)

		if gw.maxWorkers != table.Expected {
			t.Errorf("Expected %d, Got %d", table.Expected, gw.maxWorkers)
		}
	}
}

func TestBufferedQArg(t *testing.T) {
	tables := []struct {
		Given    uint32
		Expected uint32
	}{
		{defaultQSize, defaultQSize},
		{defaultQSize - 1, defaultQSize},
		{defaultQSize + 1, defaultQSize + 1},
	}

	for _, table := range tables {
		opts := Options{QSize: table.Given}
		_ = New(opts)
	}
}

func TestSubmitAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	gw.Stop()
	gw.Submit(func() {})
}

func TestStopAfterDelay(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	for gw.JobNum() != 0 {
	}
	gw.Stop()
}

func TestSubmitCheckErrorAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.SubmitCheckError(func() error {
		fn(1)
		return nil
	})

	gw.Stop()
	gw.SubmitCheckError(func() error { return nil })
}

func TestSubmitCheckResultAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.SubmitCheckResult(func() (interface{}, error) {
		fn(1)
		return nil, nil
	})

	gw.Stop()
	gw.SubmitCheckResult(func() (interface{}, error) { return nil, nil })
}

func TestSubmitCheckErrorNotSendNilToErrChan(t *testing.T) {
	gw := New()
	defer gw.Stop()

	done := make(chan struct{})
	job := func() error {
		defer close(done)
		return nil
	}

	gw.SubmitCheckError(job)
	<-done

	select {
	case err := <-gw.ErrChan:
		if err == nil {
			t.Errorf("Expected non-nil, received nil")
		}
	default:
	}
}

func TestSubmitCheckErrorUnreadChan(t *testing.T) {
	gw := New()

	for i := 0; i < 301; i++ {
		n := i
		gw.SubmitCheckError(func() error {
			if n%2 == 0 {
				return nil
			}
			return fmt.Errorf("error")
		})
	}

	gw.Stop()
}

func TestSubmitCheckResultUnreadChan(t *testing.T) {
	gw := New()

	for i := 0; i < 501; i++ {
		n := i
		gw.SubmitCheckResult(func() (interface{}, error) {
			if n%2 == 0 {
				return "output", nil
			}
			return nil, fmt.Errorf("error")
		})
	}

	gw.Stop()
}

func TestStopAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	gw.Stop()
	gw.Stop()
}

func TestLongJobs(t *testing.T) {
	gw := New()

	fn := func(i int) {
		time.Sleep(time.Duration(i) * time.Second)
	}

	for _, value := range []int{12, 12} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}

	gw.Stop()
}

func TestTimerReset(t *testing.T) {
	gw := New(Options{Workers: 1100})

	fn := func(i int) {
		time.Sleep(time.Duration(i) * time.Second)
	}

	for value := 0; value < 500; value++ {
		gw.Submit(func() {
			fn(5)
		})
	}

	gw.Stop()
}

/* ===================== Benchmarks ===================== */

func BenchmarkWithoutArgs(b *testing.B) {
	gw := New()

	for i := 0; i < b.N; i++ {
		gw.Submit(func() {})
	}

	gw.Stop()
}

func BenchmarkWithArgs(b *testing.B) {
	opts := Options{Workers: 500}
	gw := New(opts)

	for i := 0; i < b.N; i++ {
		gw.Submit(func() {})
	}

	gw.Stop()
}

func BenchmarkWithArgsError(b *testing.B) {
	opts := Options{Workers: 500}
	gw := New(opts)

	for i := 0; i < b.N; i++ {
		gw.SubmitCheckError(func() error {
			return nil
		})
	}

	gw.Stop()
}

func BenchmarkWithArgsResult(b *testing.B) {
	opts := Options{Workers: 500}
	gw := New(opts)

	for i := 0; i < b.N; i++ {
		gw.SubmitCheckResult(func() (interface{}, error) {
			return nil, nil
		})
	}

	gw.Stop()
}

/* ===================== Examples ===================== */

func Example() {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, i := range []int{9, 7, 1, 2, 3} {
		gw.Submit(func() {
			fn(i)
		})
	}

	log.Println("Submitted!")

	gw.Stop()
}

func Example_withArgs() {
	opts := Options{Workers: 3, QSize: 256}
	gw := New(opts)

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, value := range []int{9, 7, 1, 2, 3} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()
}

func Example_simple() {
	gw := New()

	gw.Submit(func() {
		fmt.Println("Hello, how are you?")
	})

	gw.Submit(func() {
		fmt.Println("I'm fine, thank you!")
	})

	log.Println("Submitted!")

	gw.Stop()
}

func Example_benchmark() {
	tStart := time.Now()

	opts := Options{Workers: 500}
	gw := New(opts)

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(5) * time.Second)
		fmt.Println("End Job", i)
	}

	for value := 500; value > 0; value-- {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	tEnd := time.Now()
	tDiff := tEnd.Sub(tStart)

	log.Println("Time taken to execute 500 jobs that are 5 seconds long is", tDiff.Seconds())
}

func Example_errorChannel() {
	gw := New()

	// You must strictly start reading from the error channel before invoking
	// SubmitCheckError() else you'll miss the updates.
	// You can employ any mechanism to read from this channel.
	go func() {
		// Error channel provides errors from job, if any
		for err := range gw.ErrChan {
			fmt.Println(err)
		}
	}()

	// This is your actual function
	fn := func(i int) error {
		// Do work here
		return fmt.Errorf("Got error %d", i)
	}

	// The job submit part
	for _, value := range []int{3, 2, 1} {
		i := value
		gw.SubmitCheckError(func() error {
			return fn(i)
		})
	}
	log.Println("Submitted!")

	// Wait for jobs to finish
	gw.Stop()
}

func Example_outputChannel() {
	gw := New()

	type myOutput struct {
		Idx  int
		Name string
	}

	// You must strictly start reading from the error and output channels
	// before invoking SubmitCheckResult() else you'll miss the updates.
	// You can employ any mechanism to read from these channels.
	go func() {
		for {
			select {
			// Error channel provides errors from job, if any
			case err := <-gw.ErrChan:
				fmt.Printf("Error: %s\n", err.Error())
			// Result channel provides output from job, if any
			// It will be of type interface{}
			case res := <-gw.ResultChan:
				fmt.Printf("Type: %T, Value: %+v\n", res, res)
			}
		}
	}()

	// This is your actual function
	fn := func(i int) (interface{}, error) {
		// Do work here

		// return error
		if i%2 == 0 {
			return nil, fmt.Errorf("Got error %d", i)
		}
		// return output
		return myOutput{Idx: i, Name: "dummy"}, nil
	}

	// The job submit part
	for _, value := range []int{3, 2, 1} {
		i := value
		gw.SubmitCheckResult(func() (interface{}, error) {
			return fn(i)
		})
	}
	log.Println("Submitted!")

	// Wait for jobs to finish
	gw.Stop()
}

func ExampleNew_withoutArgs() {
	_ = New()
}

func ExampleNew_withArgs() {
	opts := Options{Workers: 3, QSize: 256}
	_ = New(opts)
}

func ExampleGoWorkers_Submit() {
	gw := New()

	gw.Submit(func() {
		fmt.Println("Hello, how are you?")
	})

	gw.Stop()
}

func ExampleGoWorkers_SubmitCheckError() {
	gw := New()

	gw.SubmitCheckError(func() error {
		// Do some work here
		return fmt.Errorf("This is an error message")
	})

	gw.Stop()
}

func ExampleGoWorkers_SubmitCheckResult() {
	gw := New()

	gw.SubmitCheckResult(func() (interface{}, error) {
		// Do some work here
		return fmt.Sprintf("This is an output message"), nil
	})

	gw.Stop()
}
