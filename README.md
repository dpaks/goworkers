# GoWorkers
[![CircleCI](https://circleci.com/gh/dpaks/goworkers.svg?style=shield)](https://app.circleci.com/pipelines/github/dpaks/goworkers)
[![Codecov](https://codecov.io/gh/dpaks/goworkers/branch/master/graph/badge.svg)](https://codecov.io/gh/dpaks/goworkers)
[![Go Report Card](https://goreportcard.com/badge/github.com/dpaks/goworkers)](https://goreportcard.com/report/github.com/dpaks/goworkers)
[![License](https://img.shields.io/github/license/dpaks/goworkers?color=blue)](https://github.com/dpaks/goworkers/blob/master/LICENSE)

A minimal and efficient scalable workerpool implementation in Go using goroutines.

**Note:** Do not use master branch. Use the latest release.

[![GoDoc](https://godoc.org/github.com/dpaks/goworkers?status.svg)](https://godoc.org/github.com/dpaks/goworkers)

## Table of Contents
- [Installation](#installation)
- [Examples](#examples)
  - [Basic](#basic)
  - [With Arguments](#with-arguments)
  - [Without Arguments](#without-arguments)
  - [How Fast?](#benchmark)
  - [Return Error from Job](#to-receive-error-from-job)
  - [Return Output and Error from Job](#to-receive-output-and-error-from-job)
- [TODO](#todo)
- [FAQ](#faq)

## Installation
```
$ go get github.com/dpaks/goworkers
```

## Examples


###### Basic
```go
package main

import "github.com/dpaks/goworkers"

func main() {
	// initialise
	gw := goworkers.New()

	// non-blocking call
	gw.Submit(func() {
	// do your work here
	})

	// wait till your job finishes
	gw.Stop(false)
}
```

###### With arguments
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dpaks/goworkers"
)

func main() {
	opts := goworkers.Options{Workers: 20}
	gw := goworkers.New(opts)
	
	// your actual work
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
```

###### Without arguments
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dpaks/goworkers"
)

func main() {
	gw := goworkers.New()
	
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
```

###### Benchmark
```go
package main

import (
    "log"
    "time"

    "github.com/dpaks/goworkers"
)

func main() {
    tStart := time.Now()

    gw := goworkers.New()

    fn := func() {
        time.Sleep(time.Duration(5) * time.Second)
    }

    for value := 500; value > 0; value-- {
        gw.Submit(func() {
            fn()
        })
    }

    gw.Stop(false)

    tEnd := time.Now()
    tDiff := tEnd.Sub(tStart)

    log.Println("Time taken to execute 500 jobs that were 5 seconds long is only", tDiff.Seconds(), "seconds!")
}
```
**Output:** 2020/07/03 20:03:01 Time taken to execute 500 jobs that were 5 seconds long is only 5.001186599 seconds!

###### To Receive Error from Job
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dpaks/goworkers"
)

func main() {
    gw := goworkers.New()

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
```

###### To Receive Output and Error from Job
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dpaks/goworkers"
)

func main() {
    gw := goworkers.New()

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
```

## TODO
- [x] Add logs toggle
- [x] When the goworkers machine is stopped, ensure that everything is cleanedup
- [x] Add support for a 'results' channel
- [x] An option to auto-adjust worker pool size
- [ ] Introduce timeout

## FAQ

**Q.** I don't want to use error channel. I only need output. What do I?

**A.** Listen only to output channel. It is not compulsory to listen to any channel if you don't need any output.

**Q.** I get duplicate output.

**A.** In the below _wrong_ snippet, k and v are initialised only once. Since references are passed to the _Submit_ function, they may get overwritten with the newer value.

Wrong code
```go
for k, v := range myMap {
    wg.SubmitCheckResult(func() (interface{}, error) {
            return myFunc(k, v)
})
```

Correct code
```go
for i, j := range myMap {
    k := i
    v := j
    wg.SubmitCheckResult(func() (interface{}, error) {
            return myFunc(k, v)
})
```

**Q.** Can I use a combination of _Submit()_, _SubmitCheckError()_ and _SubmitCheckResult()_ and still use output and error channels?

**A.** It is absolutely safe.
