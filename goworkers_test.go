package goworkers

import (
	"fmt"
	"io/ioutil"
	"log"
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
	log.Println("Submitted!")

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
	log.Println("Submitted!")

	gw.Stop()

	if errResps != errVals {
		t.Errorf("Expected %d error responses, got %d", errVals, errResps)
	}

	<-edone
}

func TestFunctionalityCheckResultWithoutArgs(t *testing.T) {
	edone := make(chan struct{})
	rdone := make(chan struct{})
	errResps := 0
	errVals := 0
	resResps := 0
	resVals := 0
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
			errVals++
			return nil, fmt.Errorf("e%d", i)
		}
		resVals++
		return fmt.Sprintf("v%d", i), nil
	}

	for val := 0; val < rNum; val++ {
		i := val
		gw.SubmitCheckResult(func() (interface{}, error) {
			return fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	if errResps != errVals {
		t.Errorf("Expected %d error responses, got %d", errVals, errResps)
	}

	if resResps != resVals {
		t.Errorf("Expected %d result responses, got %d", resVals, resResps)
	}

	<-edone
	<-rdone
}

func TestFunctionalityWithArgs(t *testing.T) {
	opts := Options{Workers: 3}
	gw := New(opts)

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})
	log.Println("Submitted!")

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
	log.Println("Submitted!")

	gw.Stop()
	gw.Submit(func() {})
}

func TestSubmitCheckErrorAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.SubmitCheckError(func() error {
		fn(1)
		return nil
	})
	log.Println("Submitted!")

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
	log.Println("Submitted!")

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
	log.Println("Submitted!")

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
	log.Println("Submitted!")

	gw.Stop()
}

func TestStopAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
	}

	gw.Submit(func() {
		fn(1)
	})
	log.Println("Submitted!")

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
	log.Println("Submitted!")

	gw.Stop()
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
	log.Println("Submitted!")

	gw.Stop()
	gw.Stop()
}

func TestDebug(t *testing.T) {
	gw := New()
	gw.debug()
}

func TestLogsArg(t *testing.T) {
	enableLog = true
	linfo.SetOutput(ioutil.Discard)
	lerror.SetOutput(ioutil.Discard)
	ldebug.SetOutput(ioutil.Discard)
	gw := New(Options{Workers: 500})
	gw.Submit(func() { time.Sleep(15 * time.Second) })
	gw.SubmitCheckError(func() error { return nil })
	gw.SubmitCheckResult(func() (interface{}, error) { return nil, nil })
	gw.Stop()
	gw.Submit(func() {})
	gw.SubmitCheckError(func() error { return nil })
	gw.SubmitCheckResult(func() (interface{}, error) { return nil, nil })
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
