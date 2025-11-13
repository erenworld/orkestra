package task

import (
	"time"

	"github.com/google/uuid"
	"github.com/docker/go-connections/nat"
) 

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

// A Task that a user wants to run on our cluster.
// A Task can be in several states: pending, scheduled, running, completed, failed
type Task struct {
	ID			  uuid.UUID
	ContainerId	  string
	Name 		  string
	State 		  State
	Image 		  string
	Memory  	  int
	Disk 		  int
	ExposedPorts  nat.PortSet
	PortBinding   map[string]string
	RestartPolicy string
	StartTime	  time.Time
	FinishTime	  time.Time
}

// Event struct to tell the system to stop a task.
type TaskEvent struct {
	ID			uuid.UUID
	State		State		// From state one to state two
	Timestamp 	time.Time	// To record the time the event was requested
	Task 		Task
}
