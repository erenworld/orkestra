package task

import (
	"context"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
) 

type State int

const (
	Pending State = iota 	// The starting point for every task.
	Scheduled				// Once the manager has scheduled it onto a worker.
	Running					// Task moves here when a worker successfully starts the task.
	Completed				// When it completes its work in a normal way.
	Failed					// When a task fails.
)

var stateTransitionMap = map[State][]State{
	Scheduled: 	[]State{Scheduled},
	Running: 	[]State{Scheduled, Running, Failed},
	Completed: 	[]State{},
	Failed: 	[]State{},
}

// A Task that a user wants to run on our cluster (pending, scheduled, running, completed, failed).
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

type Config struct {
	Name		  string
	ContainerId	  string
	AttachStdin	  bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd			  []string
	Image		  string
	Cpu 		  float64
	Memory		  int64
	Disk		  int64
	Env			  []string
	RestartPolicy string
}

// Encapsulation to run our task as docker container.
type Docker struct {
	Client *client.Client
	Config Config
}

type DockerResult struct {
	ContainerId	string
	Error 		error
	Action 		string
	Result 		string
}

// Perform the same operations as the docker run command
func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	// Copy the stream of logs from Docker to the terminal
	io.Copy(os.Stdout, reader) 

	restartPolicy := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	resources := container.Resources{
		Memory: d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	config := container.Config{
		Image: d.Config.Image,
		Tty: false,
		Env: d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hostConfig := container.HostConfig{
		RestartPolicy: restartPolicy,
		Resources: resources,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(
		ctx, &config, &hostConfig, nil, nil, d.Config.Name,
	)
	if err != nil {
		log.Printf("Error creating container using image: %s, %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	d.Config.ContainerId = resp.ID

	out, err := d.Client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{
		ContainerId: resp.ID,
		Action: "start",
		Result: "success",
	}
}

// Perform the same operations as the docker stop command
func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Stopping container %v...\n", id)

	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, container.RemoveOptions{
		RemoveVolumes: 	true,
		RemoveLinks: 	false,
		Force:			false,
	}) 
	if err != nil {
		log.Printf("Error removing container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{
		Action: "stop",
		Result: "success",
		Error: 	nil,
	}
}

