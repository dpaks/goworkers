# GoWorkers
[![CircleCI](https://circleci.com/gh/dpaks/goworkers.svg?style=shield)](https://app.circleci.com/pipelines/github/dpaks/goworkers)
[![Codecov](https://codecov.io/gh/dpaks/goworkers/branch/master/graph/badge.svg)](https://codecov.io/gh/dpaks/goworkers)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/dpaks/goworkers/blob/master/LICENSE)

A minimal and efficient workerpool implementation in Go using goroutines.
> This project is just in alpha phase. Work commenced on 21-06-2020.

[![GoDoc](https://godoc.org/github.com/dpaks/goworkers?status.svg)](https://godoc.org/github.com/dpaks/goworkers)

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
    opts := goworkers.GoWorkersOptions{Workers: 20, Timeout: 50}
    gw := goworkers.New(opts)

    gw.Submit(func() { fmt.Println("JOB START 9"); time.Sleep(9 * time.Second); fmt.Println("JOB END 9") })
    gw.Submit(func() { fmt.Println("JOB START 7"); time.Sleep(7 * time.Second); fmt.Println("JOB END 7") })
    gw.Submit(func() { fmt.Println("JOB START 1"); time.Sleep(1 * time.Second); fmt.Println("JOB END 1") })
    gw.Submit(func() { fmt.Println("JOB START 2"); time.Sleep(2 * time.Second); fmt.Println("JOB END 2") })
    gw.Submit(func() { fmt.Println("JOB START 3"); time.Sleep(3 * time.Second); fmt.Println("JOB END 3") })
    log.Println("SUBMITTED")

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

    gw.Submit(func() { fmt.Println("JOB START 9"); time.Sleep(9 * time.Second); fmt.Println("JOB END 9") })
    gw.Submit(func() { fmt.Println("JOB START 7"); time.Sleep(7 * time.Second); fmt.Println("JOB END 7") })
    gw.Submit(func() { fmt.Println("JOB START 1"); time.Sleep(1 * time.Second); fmt.Println("JOB END 1") })
    gw.Submit(func() { fmt.Println("JOB START 2"); time.Sleep(2 * time.Second); fmt.Println("JOB END 2") })
    gw.Submit(func() { fmt.Println("JOB START 3"); time.Sleep(3 * time.Second); fmt.Println("JOB END 3") })
    log.Println("SUBMITTED")

    gw.Stop()
}
```

## TODO
- [x] Add logs toggle
- [ ] When the goworkers machine is stopped, ensure that everything is cleanedup
- [ ] Add support for a 'results' channel
- [ ] An option to auto-adjust worker pool size
