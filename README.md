# GoWorkers
[![CircleCI](https://circleci.com/gh/dpaks/goworkers.svg?style=shield)](https://app.circleci.com/pipelines/github/dpaks/goworkers)
[![Codecov](https://codecov.io/gh/dpaks/goworkers/branch/master/graph/badge.svg)](https://codecov.io/gh/dpaks/goworkers)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/dpaks/goworkers/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dpaks/goworkers)](https://goreportcard.com/report/github.com/dpaks/goworkers)

A minimal and efficient workerpool implementation in Go using goroutines.
> This project is in beta phase.
> Started on 21-06-2020.
> Do not user master branch. Pick any release, preferably the latest one.

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
	opts := goworkers.Options{Workers: 20, Timeout: 50}
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

## TODO
- [x] Add logs toggle
- [ ] When the goworkers machine is stopped, ensure that everything is cleanedup
- [ ] Add support for a 'results' channel
- [ ] An option to auto-adjust worker pool size
- [ ] Add total execution time
