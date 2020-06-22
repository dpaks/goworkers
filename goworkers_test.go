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

	gw.Submit(func() { fmt.Println("JOB START 9"); time.Sleep(9 * time.Second); fmt.Println("JOB END 9") })
	gw.Submit(func() { fmt.Println("JOB START 7"); time.Sleep(7 * time.Second); fmt.Println("JOB END 7") })
	gw.Submit(func() { fmt.Println("JOB START 1"); time.Sleep(1 * time.Second); fmt.Println("JOB END 1") })
	gw.Submit(func() { fmt.Println("JOB START 2"); time.Sleep(2 * time.Second); fmt.Println("JOB END 2") })
	gw.Submit(func() { fmt.Println("JOB START 3"); time.Sleep(3 * time.Second); fmt.Println("JOB END 3") })
	log.Println("SUBMITTED")

	gw.Stop()

	tEnd := time.Now()

	tDiff := tEnd.Sub(tStart)

	if tDiff.Seconds() > 21.0 {
		t.Errorf("Expect to complete in less than 10 seconds, took %f seconds", tDiff.Seconds())
	}
}

func TestFunctionalityWithArgs(t *testing.T) {
	tStart := time.Now()

	opts := GoWorkersOptions{Workers: 3, Logs: 1}
	gw := New(opts)

	gw.Submit(func() { fmt.Println("JOB START 9"); time.Sleep(9 * time.Second); fmt.Println("JOB END 9") })
	gw.Submit(func() { fmt.Println("JOB START 7"); time.Sleep(7 * time.Second); fmt.Println("JOB END 7") })
	gw.Submit(func() { fmt.Println("JOB START 1"); time.Sleep(1 * time.Second); fmt.Println("JOB END 1") })
	gw.Submit(func() { fmt.Println("JOB START 2"); time.Sleep(2 * time.Second); fmt.Println("JOB END 2") })
	gw.Submit(func() { fmt.Println("JOB START 3"); time.Sleep(3 * time.Second); fmt.Println("JOB END 3") })
	log.Println("SUBMITTED")

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
		opts := GoWorkersOptions{Workers: table.Given}
		gw := New(opts)

		if gw.MaxWorkerNum() != table.Expected {
			t.Errorf("Expected %d, Got %d", table.Expected, table.Given)
		}
	}
}

func Example() {
	gw := New()

	gw.Submit(func() { fmt.Println("JOB START 9"); time.Sleep(9 * time.Second); fmt.Println("JOB END 9") })
	gw.Submit(func() { fmt.Println("JOB START 7"); time.Sleep(7 * time.Second); fmt.Println("JOB END 7") })
	gw.Submit(func() { fmt.Println("JOB START 1"); time.Sleep(1 * time.Second); fmt.Println("JOB END 1") })
	gw.Submit(func() { fmt.Println("JOB START 2"); time.Sleep(2 * time.Second); fmt.Println("JOB END 2") })
	gw.Submit(func() { fmt.Println("JOB START 3"); time.Sleep(3 * time.Second); fmt.Println("JOB END 3") })
	log.Println("SUBMITTED")

	gw.Stop()
}

func ExampleNew_withoutargs() {
	_ = New()
}

func ExampleNew_withargs() {
	opts := GoWorkersOptions{Workers: 3, Logs: 1, Timeout: 20, QSize: 256}
	_ = New(opts)
}
