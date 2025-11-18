package main

import (
	"fmt"
	"orkestra/manager"
	"orkestra/node"
	"orkestra/task"
	"orkestra/worker"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name: 	"test-container-1",
		Image: 	"postgres:13",
		Env:	[]string{
			"POSTGRES_USER=orkestra",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Error creating Docker client: %v\n", err)
		return nil, &task.DockerResult{Error: err}
	}
	d := task.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run()
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, &result
	}

	fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
	return &d, &result
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf("Container %s has been stopped and removed\n", result.ContainerId)
	return &result
}

func main() {
	t := task.Task{
		ID:		uuid.New(),
		Name: 	"task-1",
		State: 	task.Pending,
		Image: 	"image-1",
		Memory: 1024,
		Disk: 	1,
	}

	te := task.TaskEvent{
		ID:			uuid.New(),
		State: 		task.Pending,
		Timestamp: 	time.Now(),
		Task:		t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)

	w := worker.Worker{
		Name: 	"worker-1",
		Queue: 	*queue.New(),
		Db:		make(map[uuid.UUID]*task.Task),
	}
	fmt.Printf("\nworker: %v\n\n", w)
	w.CollectsStats()
	w.RunTask()
	// w.StartTask()
	// w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:	 make(map[string][]*task.Task),
		EventDb: make(map[string][]*task.TaskEvent),
		Workers: []string{w.Name},
	}

	fmt.Printf("\nmanager: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWorker()

	n := node.Node{
		Name:   "linux-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 1024,
		Disk:	25,
		Role:	"worker",
	}

	fmt.Printf("\nnode: %v\n", n)

	// docker api
	fmt.Printf("Creating a container...\n")
	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("%v", createResult.Error)
		os.Exit(1)
	}
	

	time.Sleep(time.Second * 5)
	fmt.Printf("Stopping container %s\n", createResult.ContainerId)
	_ = stopContainer(dockerTask, createResult.ContainerId)
}