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
	tStart := time.Now()

	gw := New()

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

	if tDiff.Seconds() > 6.0 {
		t.Errorf("Expect to complete in less than 6 seconds, took %f seconds", tDiff.Seconds())
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

	if tDiff.Seconds() > 6.0 {
		t.Errorf("Expect to complete in less than 6 seconds, took %f seconds", tDiff.Seconds())
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

func TestDebug(t *testing.T) {
	gw := New()
	gw.debug()
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
