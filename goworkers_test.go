package goworkers

import (
	"fmt"
	"log"
	"testing"
	"time"
)

// TestStressWithoutArgs tests 500 jobs taking 5 seconds each with default 64 workers
func TestStressWithoutArgs(t *testing.T) {
	tStart := time.Now()

	opts := Options{}
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

	if tDiff.Seconds() > 41.0 {
		t.Errorf("Expect to complete in less than 6 seconds, took %f seconds", tDiff.Seconds())
	}
}

// TestStressWithArgs tests 500 jobs taking 5 seconds each with 500 workers
func TestStressWithArgs(t *testing.T) {
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

	if tDiff.Seconds() > 6.0 {
		t.Errorf("Expect to complete in less than 6 seconds, took %f seconds", tDiff.Seconds())
	}
}

func TestFunctionalityWithoutArgs(t *testing.T) {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	tStart := time.Now()
	for _, value := range []int{3, 2, 1} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	tEnd := time.Now()
	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 3.5 {
		t.Errorf("Expect to complete in less than 3.5 seconds, took %f seconds", tDiff.Seconds())
	}
}

func TestFunctionalityCheckErrorWithoutArgs(t *testing.T) {
	errResps := 0
	gw := New()

	go func() {
		for range gw.ErrChan {
			errResps++
		}
	}()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	tStart := time.Now()
	for _, value := range []int{3, 2, 1} {
		i := value
		gw.SubmitCheckError(func() error {
			fn(i)
			return nil
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	tEnd := time.Now()
	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 3.5 {
		t.Errorf("Expect to complete in less than 3.5 seconds, took %f seconds", tDiff.Seconds())
	}

	if errResps != 3 {
		t.Errorf("Expected 3 error responses, got %d", errResps)
	}
}

func TestFunctionalityCheckResultWithoutArgs(t *testing.T) {
	errResps := 0
	resResps := 0

	gw := New()

	go func() {
		for range gw.ErrChan {
			errResps++
		}
	}()

	go func() {
		for range gw.ResultChan {
			resResps++
		}
	}()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	tStart := time.Now()
	for _, value := range []int{3, 2, 1} {
		i := value
		gw.SubmitCheckResult(func() (interface{}, error) {
			fn(i)
			if i%2 == 0 {
				return nil, fmt.Errorf("error")
			}
			return "value", nil
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	tEnd := time.Now()
	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 3.5 {
		t.Errorf("Expect to complete in less than 3.5 seconds, took %f seconds", tDiff.Seconds())
	}

	if errResps != 1 {
		t.Errorf("Expected 1 error responses, got %d", errResps)
	}

	if resResps != 2 {
		t.Errorf("Expected 2 result responses, got %d", resResps)
	}
}

func TestFunctionalityWithArgs(t *testing.T) {
	tStart := time.Now()

	opts := Options{Workers: 3, Logs: 1}
	gw := New(opts)

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, value := range []int{3, 2, 1} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	tEnd := time.Now()

	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 3.5 {
		t.Errorf("Expect to complete in less than 3.5 seconds, took %f seconds", tDiff.Seconds())
	}
}

func TestWorkerArg(t *testing.T) {
	tables := []struct {
		Given    uint32
		Expected uint32
	}{
		{defaultWorkers - 1, defaultWorkers},
		{defaultWorkers + 1, defaultWorkers + 1},
	}

	for _, table := range tables {
		opts := Options{Workers: table.Given}
		gw := New(opts)

		if gw.maxWorkers != table.Expected {
			t.Errorf("Expected %d, Got %d", table.Expected, gw.maxWorkers)
		}
	}
}

func TestTimeoutArg(t *testing.T) {
	tables := []struct {
		Given    uint32
		Expected uint32
	}{
		{defaultTimeout, defaultTimeout},
		{defaultTimeout - 1, defaultTimeout},
		{defaultTimeout + 1, defaultTimeout + 1},
	}

	for _, table := range tables {
		opts := Options{Timeout: table.Given}
		gw := New(opts)

		if gw.timeout != (time.Second * time.Duration(table.Expected)) {
			t.Errorf("Expected %d, Got %d", table.Expected, gw.timeout)
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

func TestLogsArg(t *testing.T) {
	for _, logLvl := range []uint8{1, 2, 3, 0} {
		opts := Options{Logs: logLvl}
		_ = New(opts)
	}
}

func TestSubmitAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, value := range []int{2, 1} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()
	gw.Submit(func() {})
}

func TestSubmitCheckErrorAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, value := range []int{2, 1} {
		i := value
		gw.SubmitCheckError(func() error {
			fn(i)
			return nil
		})
	}
	log.Println("Submitted!")

	gw.Stop()
	gw.SubmitCheckError(func() error { return nil })
}

func TestSubmitCheckResultAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, value := range []int{2, 1} {
		i := value
		gw.SubmitCheckResult(func() (interface{}, error) {
			fn(i)
			return nil, nil
		})
	}
	log.Println("Submitted!")

	gw.Stop()
	gw.SubmitCheckResult(func() (interface{}, error) { return nil, nil })
}

func TestStopAfterStop(t *testing.T) {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for _, value := range []int{2, 1} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()
	gw.Stop()
}

func TestLongJobs(t *testing.T) {
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
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
	gw := New()

	fn := func(i int) {
		fmt.Println("Start Job", i)
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Println("End Job", i)
	}

	for value := 0; value < 500; value++ {
		gw.Submit(func() {
			fn(2)
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
	opts := Options{Workers: 3, Logs: 1, Timeout: 10, QSize: 256}
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
	opts := Options{Workers: 3, Logs: 1, Timeout: 20, QSize: 256}
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
