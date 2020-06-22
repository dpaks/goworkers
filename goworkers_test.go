package goworkers

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestFunctionalityWithoutArgs(t *testing.T) {
	tStart := time.Now()

	gw := New()

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

	tEnd := time.Now()

	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 21.0 {
		t.Errorf("Expect to complete in less than 10 seconds, took %f seconds", tDiff.Seconds())
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

	for _, value := range []int{9, 7, 1, 2, 3} {
		i := value
		gw.Submit(func() {
			fn(i)
		})
	}
	log.Println("Submitted!")

	gw.Stop()

	tEnd := time.Now()

	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 21.0 {
		t.Errorf("Expect to complete in less than 10 seconds, took %f seconds", tDiff.Seconds())
	}
}

func TestArgs(t *testing.T) {
	tables := []struct {
		Given    uint32
		Expected uint32
	}{
		{1, 2},
		{3, 3},
	}

	for _, table := range tables {
		opts := Options{Workers: table.Given}
		gw := New(opts)

		if gw.MaxWorkerNum() != table.Expected {
			t.Errorf("Expected %d, Got %d", table.Expected, table.Given)
		}
	}
}

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
	opts := Options{Workers: 3, Logs: 2, Timeout: 10, QSize: 256}
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
