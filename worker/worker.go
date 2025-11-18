package worker

import (
	"fmt"
	"log"
	"orkestra/task"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

// Run tasks as Docker containers.
// Accept tasks to run from a manager.
// Provide relevant statistics to the manager for the purpose of scheduling tasks.
// Keep track of its tasks and their state.

// The worker maintains the state of its tasks by storing them in a database.
// api -> task queue -> metrics -> runtime (docker start and stop) -> maintain state of its task by storing them in a db

type Worker struct {
	Name 		string
	Queue   	queue.Queue					// a map of UUIDs to tasks
	Db			map[uuid.UUID]*task.Task   	// Tasks are handled in FIFO order  
	TaskCount 	int
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	d := task.NewDocker(config)

	result := d.Stop(t.ContainerId)
	if result.Error != nil {
		log.Printf("Error stopping container %v, %v\n", t.ContainerId, result.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stop and removed container %v for task %v\n", t.ContainerId, t.ID)

	return result
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := task.NewConfig(&t)
	d := task.NewDocker(config)
	result := d.Run()

	if result.Error != nil {
		log.Printf("Error running task %v, %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerId = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t
	return result
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) RunTask() task.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Printf("Queue is empty")
		return task.DockerResult{Error: nil}
	}

	taskQueued := t.(task.Task)				// convert to the proper type
	taskPersisted := w.Db[taskQueued.ID]	
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	var result task.DockerResult
	if task.validStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			result := w.StartTask(taskQueued)
		case task.Completed:
			result := w.StopTask(taskQueued)
		default:
			result.Error = errors.New("We should not get here")
		}
	} else {
		err := fmt.Errorf("Invalid transition from %v to %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}

	return result
}


func (w *Worker) CollectsStats() {
	fmt.Println("Collect stats")
}