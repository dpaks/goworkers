/*
Copyright 2020 Deepak S<deepaks@outlook.in>
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

	gw.Stop(false)
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

	gw.Stop(true)

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

	gw.Stop(true)

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

	gw1.Stop(true)
	gw2.Stop(true)

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

	gw.Stop(false)
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

	gw.Stop(false)
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
	gw.Stop(false)
}

func TestWait(t *testing.T) {
	gw := New()

	fn := func() {
		time.Sleep(1 * time.Second)
	}

	for i := 0; i < 100; i++ {
		gw.Submit(func() {
			fn()
		})
	}

	if gw.JobNum() == 0 {
		t.Errorf("Number of jobs must be greater than 0")
	}

	gw.Wait(false)

	if gw.JobNum() != 0 {
		t.Errorf("Number of jobs should be 0. Got %d", gw.JobNum())
	}

	for i := 0; i < 100; i++ {
		gw.Submit(func() {
			fn()
		})
	}

	if gw.JobNum() == 0 {
		t.Errorf("Number of jobs must be greater than 0")
	}

	gw.Wait(true)

	if gw.JobNum() != 0 {
		t.Errorf("Number of jobs should be 0. Got %d", gw.JobNum())
	}

	gw.Stop(false)
}

func TestWaitAfterWait(t *testing.T) {
	gw := New()
	defer gw.Stop(false)

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	go gw.Wait(false)
	gw.Wait(false)
}

func TestSubmitCheckErrorAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.SubmitCheckError(func() error {
		fn(1)
		return nil
	})

	gw.Stop(false)
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

	gw.Stop(false)
	gw.SubmitCheckResult(func() (interface{}, error) { return nil, nil })
}

func TestSubmitCheckResultAfterStopWait(t *testing.T) {
	gw := New()

	go func() {
		for {
			select {
			case _, ok := <-gw.ResultChan:
				if !ok {
					return
				}
			case <-gw.ErrChan:
			}
		}
	}()

	fn := func(i int) {
	}

	gw.SubmitCheckResult(func() (interface{}, error) {
		fn(1)
		return nil, nil
	})

	gw.Stop(true)
	gw.SubmitCheckResult(func() (interface{}, error) { return nil, nil })
}

func TestSubmitCheckErrorNotSendNilToErrChan(t *testing.T) {
	gw := New()
	defer gw.Stop(true)

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

	gw.Stop(false)
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

	gw.Stop(false)
}

func TestStopAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})

	gw.Stop(false)
	gw.Stop(false)
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

	gw.Stop(false)
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

	gw.Stop(false)
}

/* ===================== Benchmarks ===================== */

func BenchmarkWithoutArgs(b *testing.B) {
	gw := New()

	for i := 0; i < b.N; i++ {
		gw.Submit(func() {})
	}

	gw.Stop(false)
}

func BenchmarkWithArgs(b *testing.B) {
	opts := Options{Workers: 500}
	gw := New(opts)

	for i := 0; i < b.N; i++ {
		gw.Submit(func() {})
	}

	gw.Stop(false)
}

func BenchmarkWithArgsError(b *testing.B) {
	opts := Options{Workers: 500}
	gw := New(opts)

	for i := 0; i < b.N; i++ {
		gw.SubmitCheckError(func() error {
			return nil
		})
	}

	gw.Stop(false)
}

func BenchmarkWithArgsResult(b *testing.B) {
	opts := Options{Workers: 500}
	gw := New(opts)

	for i := 0; i < b.N; i++ {
		gw.SubmitCheckResult(func() (interface{}, error) {
			return nil, nil
		})
	}

	gw.Stop(false)
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

	gw.Stop(false)
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

	gw.Stop(false)
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

	gw.Stop(false)
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

	gw.Stop(false)

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
	// Here, wait flag is set to true. Setting wait to true ensures that
	// the output channels are read from completely.
	// Stop(true) exits only when the error channel is completely read from.
	gw.Stop(true)
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
			case err, ok := <-gw.ErrChan:
				// The error channel is closed when the workers are done with their tasks.
				// When the channel is closed, ok is set to false
				if !ok {
					return
				}
				fmt.Printf("Error: %s\n", err.Error())
			// Result channel provides output from job, if any
			// It will be of type interface{}
			case res, ok := <-gw.ResultChan:
				// The result channel is closed when the workers are done with their tasks.
				// When the channel is closed, ok is set to false
				if !ok {
					return
				}
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
	// Here, wait flag is set to true. Setting wait to true ensures that
	// the output channels are read from completely.
	// Stop(true) exits only when both the result and the error channels are completely read from.
	gw.Stop(true)
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

	gw.Stop(false)
}

func ExampleGoWorkers_SubmitCheckError() {
	gw := New()

	gw.SubmitCheckError(func() error {
		// Do some work here
		return fmt.Errorf("This is an error message")
	})

	gw.Stop(true)
}

func ExampleGoWorkers_SubmitCheckResult() {
	gw := New()

	gw.SubmitCheckResult(func() (interface{}, error) {
		// Do some work here
		return fmt.Sprintf("This is an output message"), nil
	})

	gw.Stop(true)
}

func ExampleGoWorkers_Wait() {
	gw := New()
	defer gw.Stop(false)

	gw.Submit(func() {
		fmt.Println("Hello, how are you?")
	})

	gw.Wait(false)

	gw.Submit(func() {
		fmt.Println("I'm good, thank you!")
	})

	gw.Wait(false)
}
