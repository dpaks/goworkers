package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

const (
	DEFAULT_TIMEOUT = 10
	DEFAULT_WORKERS = 2
	MAX_WORKERS = 128
	MAXQ            = 100
)

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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("GoWorkers:")
}

// Creates a new worker pool with the specified number of worker
// goroutines. Minimum workers is 2.
func New(args ...int) *GoWorkers {
	gw := &GoWorkers{
		workerQ:   make(chan func()),
		jobQ:      make(chan func()),
		bufferedQ: make(chan func(), MAXQ),
		terminate: make(chan struct{}),
	}
	gw.maxWorkers = DEFAULT_WORKERS
	gw.timeout = time.Second * DEFAULT_TIMEOUT
	if len(args) == 1 && args[0] > 2 {
		gw.maxWorkers = uint32(args[0])
		log.Printf("DEBUG: %+v\n", *gw)
	}
	if len(args) == 2 && args[1] == 0 {
		gw.timeout = time.Second * DEFAULT_TIMEOUT
		log.Printf("DEBUG: %+v\n", *gw)
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

func (gw *GoWorkers) Submit(job func()) {
	if atomic.LoadInt32(&gw.stopping) == 1 {
		log.Println("Cannot accept jobs - Shutting down the go workers!")
		return
	}
	atomic.AddUint32(&gw.qnumJobs, 1)
	gw.jobQ <- job
}

func sleep(n int) {
	time.Sleep(time.Duration(n) * time.Second)
}

func (gw *GoWorkers) Stop() {
	if !atomic.CompareAndSwapInt32(&gw.stopping, 0, 1) {
		log.Println("Stop already triggered")
		return
	}
	log.Println("Requesting shut down of the go workers!")
	for {
		if ok := gw.wait(); ok {
			break
		}
	}
	log.Println("Successfully shut the go workers!")
}

func (gw *GoWorkers) wait() bool {
	if atomic.LoadInt32(&gw.stopping) == 0 {
		log.Printf("DEBUG: %+v\n", *gw)
		if gw.JobNum() != 0 || gw.QueuedJobNum() != 0 {
			log.Printf("Cannot stop. Active Jobs = %d, Queued Jobs = %d\n", gw.JobNum(), gw.QueuedJobNum())
			return false
		}
		log.Printf("DEBUG: %+v\n", *gw)
		gw.terminate <- struct{}{}
		close(gw.jobQ)
	} else if gw.JobNum() == 0 && gw.QueuedJobNum() == 0 {
		return true
	}

	return false
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
				// There shouldn't be a situation where the job is neither in active or queue state
				atomic.AddUint32(&gw.numJobs, uint32(1))
				atomic.AddUint32(&gw.qnumJobs, ^uint32(0))
				log.Printf("DEBUG: %+v\n", *gw)
			}(job)
		}
	}
}

func (gw *GoWorkers) startWorker() {
	defer func() {
		atomic.AddUint32(&gw.numWorkers, ^uint32(0))
		log.Println("Stopped idle worker. Worker count =", gw.numWorkers)
	}()

	atomic.AddUint32(&gw.numWorkers, 1)
	log.Println("Started worker. Worker count =", gw.numWorkers)
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
				log.Println("Timed out - killing self!")
				log.Printf("DEBUG: %+v\n", *gw)
				return
			}
		}
	}
}

func main() {
	gw := New(20)

	gw.Submit(func() { fmt.Println("JOB START 9"); time.Sleep(9 * time.Second); fmt.Println("JOB END 9") })
	gw.Submit(func() { fmt.Println("JOB START 7"); time.Sleep(7 * time.Second); fmt.Println("JOB END 7") })
	gw.Submit(func() { fmt.Println("JOB START 1"); time.Sleep(1 * time.Second); fmt.Println("JOB END 1") })
	gw.Submit(func() { fmt.Println("JOB START 2"); time.Sleep(2 * time.Second); fmt.Println("JOB END 2") })
	gw.Submit(func() { fmt.Println("JOB START 3"); time.Sleep(3 * time.Second); fmt.Println("JOB END 3") })
	log.Println("SUBMITTED")
	gw.Stop()
	//sleep(10)
}
