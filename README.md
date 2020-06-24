# GoWorkers
[![CircleCI](https://circleci.com/gh/dpaks/goworkers.svg?style=shield)](https://app.circleci.com/pipelines/github/dpaks/goworkers)
[![Codecov](https://codecov.io/gh/dpaks/goworkers/branch/master/graph/badge.svg)](https://codecov.io/gh/dpaks/goworkers)
[![Go Report Card](https://goreportcard.com/badge/github.com/dpaks/goworkers)](https://goreportcard.com/report/github.com/dpaks/goworkers)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/dpaks/goworkers/blob/master/LICENSE)

A minimal and efficient scalable workerpool implementation in Go using goroutines.

**Note:** Do not user master branch. Use the latest release.

[![GoDoc](https://godoc.org/github.com/dpaks/goworkers?status.svg)](https://godoc.org/github.com/dpaks/goworkers)

## Table of Contents
- [Installation](#installation)
- [Examples](#examples)
  - [With Arguments](#with-arguments)
  - [Without Arguments](#without-arguments)
  - [Benchmark](#benchmark)
  - [Return Error from Job](#to-receive-error-from-job)
  - [Return Output and Error from Job](#to-receive-output-and-error-from-job)
- [TODO](#todo)
- [FAQ](#faq)

## Installation
```
$ go get github.com/dpaks/goworkers
```

## Examples

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
	
	gw.Stop()
}
```

###### Benchmark
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dpaks/goworkers"
)

func main() {
	tStart := time.Now()

	opts := goworkers.Options{Workers: 500}
	gw := goworkers.New(opts)

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
```
**Output:** Time taken to execute 500 jobs that are 5 seconds long is 5.01778295

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
    gw.Stop()
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
```

## TODO
- [x] Add logs toggle
- [x] When the goworkers machine is stopped, ensure that everything is cleanedup
- [x] Add support for a 'results' channel
- [ ] An option to auto-adjust worker pool size
- [ ] Add total execution time

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
