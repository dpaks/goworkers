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
	defaultTimeout = 10
	// Number of initial workers spawned if unspecified
	defaultWorkers = 64
	// The size of the buffered queue where jobs are queued up if no
	// workers are available to process the incoming jobs, unless specified
	defaultQSize = 128
)

var (
	// Package level loggers
	linfo  *log.Logger
	ldebug *log.Logger
	lerror *log.Logger
)

// GoWorkers is a collection of worker goroutines.
//
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

// Options configures the behaviour of worker pool.
//
// Timeout specifies the time after which an idle worker goroutine will be killed.
// Default timeout is 10 seconds.
//
// Workers specifies the number of workers that will be spawned.
// If unspecified, workers will be progressively spawned upto a maximum of 64.
//
// Logs accepts log levels - 0 (default), 1, 2.
// 0: Only error logs.
// 1: Only info and error logs.
// 2: lerror, info and debug logs (badly verbose).
//
// QSize specifies the size of the queue that holds up incoming jobs.
// Minimum value is 128.
type Options struct {
	Timeout uint32
	Workers uint32
	Logs    uint8
	QSize   uint32
}

func init() {
	linfo = log.New(ioutil.Discard, "(GoWorkers)INFO: ", log.LstdFlags|log.Lshortfile)
	ldebug = log.New(ioutil.Discard, "(GoWorkers)DEBUG: ", log.LstdFlags|log.Lshortfile)
	lerror = log.New(os.Stderr, "(GoWorkers)ERROR: ", log.LstdFlags|log.Lshortfile)
}

// New creates a new worker pool.
//
// Accepts optional Options{} argument.
func New(args ...Options) *GoWorkers {
	gw := &GoWorkers{
		workerQ: make(chan func()),
		// Do not remove jobQ - to stop receiving input once Stop() is called
		jobQ:      make(chan func()),
		terminate: make(chan struct{}),
	}

	gw.maxWorkers = defaultWorkers
	gw.timeout = time.Second * defaultTimeout
	gw.bufferedQ = make(chan func(), defaultQSize)
	if len(args) == 1 {
		if args[0].Workers > defaultWorkers {
			gw.maxWorkers = args[0].Workers
		}
		if args[0].Timeout > defaultTimeout {
			gw.timeout = time.Second * time.Duration(args[0].Timeout)
		}
		if args[0].Logs == 1 {
			linfo.SetOutput(os.Stdout)
		} else if args[0].Logs == 2 {
			linfo.SetOutput(os.Stdout)
			ldebug.SetOutput(os.Stdout)
		}
		if args[0].QSize > defaultQSize {
			gw.bufferedQ = make(chan func(), args[0].QSize)
		}
	}

	go gw.start()

	return gw
}

// JobNum returns number of active jobs
func (gw *GoWorkers) JobNum() uint32 {
	return atomic.LoadUint32(&gw.numJobs)
}

// QueuedJobNum returns number of queued jobs
func (gw *GoWorkers) QueuedJobNum() uint32 {
	return atomic.LoadUint32(&gw.qnumJobs)
}

// WorkerNum returns number of active workers
func (gw *GoWorkers) WorkerNum() uint32 {
	return atomic.LoadUint32(&gw.numWorkers)
}

// MaxWorkerNum returns maximum number of workers
func (gw *GoWorkers) MaxWorkerNum() uint32 {
	return atomic.LoadUint32(&gw.maxWorkers)
}

// Submit is a non-blocking call with arg of type func()
func (gw *GoWorkers) Submit(job func()) {
	if atomic.LoadInt32(&gw.stopping) == 1 {
		lerror.Println("Cannot accept jobs - Shutting down the go workers!")
		return
	}
	atomic.AddUint32(&gw.qnumJobs, 1)
	gw.jobQ <- job
}

func msleep(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}

// Stop gracefully waits for jobs to finish running.
//
// This is a non-blocking call and returns when all the active and queued jobs are finished.
func (gw *GoWorkers) Stop() {
	if !atomic.CompareAndSwapInt32(&gw.stopping, 0, 1) {
		linfo.Println("Stop already triggered")
		return
	}
	linfo.Println("Requesting shut down of the go workers!")
	for {
		if gw.JobNum() != 0 || gw.QueuedJobNum() != 0 {
			ldebug.Printf("Cannot stop. Active Jobs = %d, Queued Jobs = %d\n", gw.JobNum(), gw.QueuedJobNum())
			msleep(500)
			continue
		}
		gw.terminate <- struct{}{}
		// close the input channel
		close(gw.jobQ)
		break
	}
	linfo.Println("Successfully shut the go workers!")
}

func (gw *GoWorkers) debug() {
	ldebug.Printf("\n***\n numWorkers: %d \n maxWorkers: %d\n numJobs: %d\n qnumJobs: %d\n timeout: %f\n stopping: %d\n***\n",
		gw.numWorkers, gw.maxWorkers, gw.numJobs, gw.qnumJobs, gw.timeout.Seconds(), gw.stopping)
}

func (gw *GoWorkers) start() {
	defer func() {
		close(gw.bufferedQ)
		close(gw.workerQ)
	}()
	for {
		select {
		case <-gw.terminate:
			return
		case job := <-gw.jobQ:
			gw.bufferedQ <- job
		}

		select {
		case job := <-gw.bufferedQ:
			go func(job func()) {
				// Should be ideally an atomic operation. However, an extra goroutine tradesoff
				// better than using a mutex.
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
		linfo.Println("Stopped idle worker. Worker count =", gw.numWorkers)
	}()

	atomic.AddUint32(&gw.numWorkers, 1)
	linfo.Println("Started worker. Worker count =", gw.numWorkers)
	timer := time.NewTimer(gw.timeout)

	for {
		select {
		case job, ok := <-gw.workerQ:
			if !ok {
				return
			}
			job()
			atomic.AddUint32(&gw.numJobs, ^uint32(0))
			timer.Reset(gw.timeout)
		case <-timer.C:
			// Should be ideally an atomic operation. However, an extra goroutine tradesoff
			// better than using a mutex.
			if (gw.JobNum() + gw.QueuedJobNum()) < gw.WorkerNum() {
				linfo.Println("Timed out - killing self!")
				return
			}
		}
	}
}
