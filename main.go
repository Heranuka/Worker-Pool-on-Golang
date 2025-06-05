package main

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	id       int
	DoneChan chan struct{}
	wg       *sync.WaitGroup
}

func NewWorker(id int, wg *sync.WaitGroup, consumer <-chan string) *Worker {
	pool := &Worker{
		id:       id,
		DoneChan: make(chan struct{}),
		wg:       wg,
	}
	go pool.Start(consumer)

	return pool
}

func (w *Worker) Start(consumer <-chan string) {
	defer w.wg.Done()

	for {
		select {
		case task, ok := <-consumer:
			if !ok {
				return
			}
			fmt.Printf("Worker %d: started %s\n", w.id, task)
			time.Sleep(time.Millisecond * 500)
			fmt.Printf("Worker %d: finished %s\n", w.id, task)
		case <-w.DoneChan:
			return
		}
	}
}

func (w *Worker) Stop() {
	close(w.DoneChan)
}

func main() {
	var wg sync.WaitGroup
	numWorkers := 5
	numTasks := 15
	inputChan := make(chan string, numTasks)
	workers := make(map[int]*Worker)
	AddWorkerChan := make(chan struct{})
	DeleteWorkerChan := make(chan struct{})

	var workerMutex sync.Mutex
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		workers[i] = NewWorker(i, &wg, inputChan)
	}

	go func() {
		for j := 1; j <= numTasks; j++ {
			inputChan <- fmt.Sprintf("Task %d", j)
			time.Sleep(time.Millisecond * 500)
		}
		close(inputChan)
	}()

	// Динамически добавляем воркеров
	go func() {
		<-AddWorkerChan
		AddWorker(3, &numWorkers, &wg, workers, inputChan, &workerMutex)
	}()

	// Динамически удаляем воркеров
	go func() {
		<-DeleteWorkerChan
		DeleteWorker(1, &workers, &workerMutex)

	}()

	time.Sleep(2 * time.Second)
	AddWorkerChan <- struct{}{}
	time.Sleep(2 * time.Second)
	DeleteWorkerChan <- struct{}{}

	wg.Wait()

	fmt.Println("All workers finished their tasks")
}

func AddWorker(amount int, numWorkers *int, wg *sync.WaitGroup, workers map[int]*Worker, inputChan <-chan string, workerMutex *sync.Mutex) {
	for i := 0; i < amount; i++ {
		workerMutex.Lock()
		*numWorkers++
		wg.Add(1)
		workers[*numWorkers] = NewWorker(*numWorkers, wg, inputChan)
		workerMutex.Unlock()
		fmt.Printf("%d amount of workers were added to stream\n", amount)
		time.Sleep(time.Second) // Пауза между добавлением воркеров

	}
}

func DeleteWorker(id int, workers *map[int]*Worker, workerMutex *sync.Mutex) {
	workerMutex.Lock()
	defer workerMutex.Unlock()
	if worker, exists := (*workers)[id]; exists {
		worker.Stop()
		delete(*workers, id)
		fmt.Printf("Worker with %d id was deleted from stream\n", id)
	}
}
