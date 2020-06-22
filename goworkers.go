// Package goworkers implements a simple, flexible and lightweight
// goroutine worker pool implementation.
package goworkers

import (
	"io/ioutil"
	"log"
	"os"
	"sync/atomic"
	"time"
)

const (
	// Time after which a worker will be killed if inactive
	DEFAULT_TIMEOUT = 10
	// Number of initial workers spawned if unspecified
	DEFAULT_WORKERS = 2
	// The size of the buffered queue where jobs are queued up if no
	// workers are available to process the incoming jobs, unless specified
	DEFAULT_QSIZE = 100
)

var (
	Info  *log.Logger
	Debug *log.Logger
	Error *log.Logger
)

// GoWorkers is a collection of worker goroutines.
// Idle workers will be timed out. At minimum, 2 workers will be spawned.
type GoWorkers struct {
	numWorkers uint32
	maxWorkers uint32
	numJobs    uint32
	qnumJobs   uint32
	timeout    time.Duration
	workerQ    chan func()
	bufferedQ  chan func()
	jobQ       chan func()
	terminate  chan struct{}
	stopping   int32
}

// Options to configure the behaviour of worker pool.
// Timeout specifies the time after which an idle worker goroutine will be killed.
// Default timeout is DEFAULT_TIMEOUT seconds.
// Workers specifies the number of workers that will be spawned.
// Default number of workers is DEFAULT_WORKERS.
// Logs accepts log levels - 0 (default), 1, 2
// Log level 0: Only error logs.
// Log level 1: Only info and error logs.
// Log level 2: Error, info and debug logs. (badly verbose)
// QSize specifies the size of the queue that holds up incoming jobs.
// Minimum value is DEFAULT_QSIZE
type GoWorkersOptions struct {
	Timeout uint32
	Workers uint32
	Logs    uint8
	QSize   uint32
}

func init() {
	Info = log.New(ioutil.Discard, "(GoWorkers)INFO: ", log.LstdFlags|log.Lshortfile)
	Debug = log.New(ioutil.Discard, "(GoWorkers)DEBUG: ", log.LstdFlags|log.Lshortfile)
	Error = log.New(os.Stderr, "(GoWorkers)ERROR: ", log.LstdFlags|log.Lshortfile)
}

// Creates a new worker pool.
// Accepts optional GoWorkersOptions{} argument.
func New(args ...GoWorkersOptions) *GoWorkers {
	gw := &GoWorkers{
		workerQ:   make(chan func()),
		jobQ:      make(chan func()),
		terminate: make(chan struct{}),
	}

	gw.maxWorkers = DEFAULT_WORKERS
	gw.timeout = time.Second * DEFAULT_TIMEOUT
	gw.bufferedQ = make(chan func(), DEFAULT_QSIZE)
	if len(args) == 1 {
		if args[0].Workers > DEFAULT_WORKERS {
			gw.maxWorkers = args[0].Workers
		}
		if args[0].Timeout > DEFAULT_TIMEOUT {
			gw.timeout = time.Second * time.Duration(args[0].Timeout)
		}
		if args[0].Logs == 1 {
			Info.SetOutput(os.Stdout)
		} else if args[0].Logs == 2 {
			Info.SetOutput(os.Stdout)
			Debug.SetOutput(os.Stdout)
		}
		if args[0].QSize > DEFAULT_QSIZE {
			gw.bufferedQ = make(chan func(), args[0].QSize)
		}
	}

	go gw.start()

	return gw
}

// Number of active jobs
func (gw *GoWorkers) JobNum() uint32 {
	return atomic.LoadUint32(&gw.numJobs)
}

// Number of queued jobs
func (gw *GoWorkers) QueuedJobNum() uint32 {
	return atomic.LoadUint32(&gw.qnumJobs)
}

// Number of active workers
func (gw *GoWorkers) WorkerNum() uint32 {
	return atomic.LoadUint32(&gw.numWorkers)
}

// Maximum number of workers
func (gw *GoWorkers) MaxWorkerNum() uint32 {
	return atomic.LoadUint32(&gw.maxWorkers)
}

// Non-blocking call to submit jobs of type job()
func (gw *GoWorkers) Submit(job func()) {
	if atomic.LoadInt32(&gw.stopping) == 1 {
		Error.Println("Cannot accept jobs - Shutting down the go workers!")
		return
	}
	atomic.AddUint32(&gw.qnumJobs, 1)
	gw.jobQ <- job
}

func msleep(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}

// Gracefully waits for jobs to finish running.
// This is a non-blocking call and returns when all the active and queued jobs are finished.
func (gw *GoWorkers) Stop() {
	if !atomic.CompareAndSwapInt32(&gw.stopping, 0, 1) {
		Info.Println("Stop already triggered")
		return
	}
	Info.Println("Requesting shut down of the go workers!")
	for {
		if gw.JobNum() != 0 || gw.QueuedJobNum() != 0 {
			Debug.Printf("Cannot stop. Active Jobs = %d, Queued Jobs = %d\n", gw.JobNum(), gw.QueuedJobNum())
			msleep(500)
			continue
		}
		gw.terminate <- struct{}{}
		close(gw.jobQ)
		break
	}
	Info.Println("Successfully shut the go workers!")
}

func (gw *GoWorkers) start() {
	for {
		select {
		case <-gw.terminate:
			return
		case job, ok := <-gw.jobQ:
			if !ok {
				continue
			}
			gw.bufferedQ <- job
		}

		select {
		case job, ok := <-gw.bufferedQ:
			if !ok {
				continue
			}

			go func(job func()) {
				if (gw.WorkerNum() < 2) || (gw.WorkerNum() < gw.MaxWorkerNum() && gw.QueuedJobNum() >= 1) {
					go gw.startWorker()
				}
				gw.workerQ <- job
				// Move job to active before removing from queue
				// There shouldn't be a situation where the job is neither in active nor in queue state
				atomic.AddUint32(&gw.numJobs, uint32(1))
				atomic.AddUint32(&gw.qnumJobs, ^uint32(0))
			}(job)
		}
	}
}

func (gw *GoWorkers) startWorker() {
	defer func() {
		atomic.AddUint32(&gw.numWorkers, ^uint32(0))
		Info.Println("Stopped idle worker. Worker count =", gw.numWorkers)
	}()

	atomic.AddUint32(&gw.numWorkers, 1)
	Info.Println("Started worker. Worker count =", gw.numWorkers)
	timer := time.NewTimer(gw.timeout)

	for {
		select {
		case job, ok := <-gw.workerQ:
			if !ok {
				continue
			}
			if job == nil {
				return
			}

			job()
			atomic.AddUint32(&gw.numJobs, ^uint32(0))
			timer.Reset(gw.timeout)
		case <-timer.C:
			if (gw.JobNum() + gw.QueuedJobNum()) < gw.WorkerNum() {
				Info.Println("Timed out - killing self!")
				return
			}
		}
	}
}
