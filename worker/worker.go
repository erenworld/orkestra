package worker

import (
	"fmt"
	"orkestra/task"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

// Run tasks as Docker containers
// Accept tasks to run from a manager
// Provide relevant statistics to the manager for the purpose of scheduling tasks
// Keep track of its tasks and their state

// TODOS: implement my own queue in Golang
// TODOS: implement my own DB
type Worker struct {
	Name 		string
	Queue   	queue.Queue					// a map of UUIDs to tasks
	Db			map[uuid.UUID]*task.Task   	// Tasks are handled in FIFO order  
	TaskCount 	int
}

// Handle running a task on the machine where the worker is running.
func (w *Worker) RunTask() {
	fmt.Println("Stats")
}

func (w *Worker) StartTask() {
	fmt.Println("Stats")
}

func (w *Worker) StopTask() {
	fmt.Println("Stats")
}

func (w *Worker) CollectsStats() {
	fmt.Println("Stats")
}