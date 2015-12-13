package main

import (
	"flag"
	"fmt"
	"time"

	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type Executor struct {
	tasksLaunched int
}

func newExampleExecutor() *Executor {
	return &Executor{tasksLaunched: 0}
}

func (exec *Executor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	fmt.Println("Registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *Executor) Reregistered(driver exec.ExecutorDriver, slaveInfo *mesos.SlaveInfo) {
	fmt.Println("Re-registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *Executor) Disconnected(exec.ExecutorDriver) {
	fmt.Println("Executor disconnected.")
}

func (exec *Executor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	fmt.Printf("Launching task %v with data [%#x]\n", taskInfo.GetName(), taskInfo.Data)

	runStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_RUNNING.Enum(),
	}
	_, err := driver.SendStatusUpdate(runStatus)
	if err != nil {
		fmt.Println("Got error", err)
	}

	exec.tasksLaunched++
	fmt.Println("Total tasks launched ", exec.tasksLaunched)

	fmt.Println("I would be doing work here.... instead ill sleep for 30s")
	time.Sleep(30 * time.Second)
	/*
		// Download image
		fileName, err := downloadImage(string(taskInfo.Data))
		if err != nil {
			fmt.Printf("Failed to download image with error: %v\n", err)
			return
		}
		fmt.Printf("Downloaded image: %v\n", fileName)

		// Process image
		fmt.Printf("Processing image: %v\n", fileName)
		outFile, err := procImage(fileName)
		if err != nil {
			fmt.Printf("Failed to process image with error: %v\n", err)
			return
		}

		// Upload image
		fmt.Printf("Uploading image: %v\n", outFile)
		if err = uploadImage("http://127.0.0.1:12345/", outFile); err != nil {
			fmt.Printf("Failed to upload image with error: %v\n", err)
			return
		} else {
			fmt.Printf("Uploaded image: %v\n", outFile)
		}
	*/

	// Finish task
	fmt.Println("Finishing task", taskInfo.GetName())
	finStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_FINISHED.Enum(),
	}
	_, err = driver.SendStatusUpdate(finStatus)
	if err != nil {
		fmt.Println("Got error", err)
		return
	}

	fmt.Println("Task finished", taskInfo.GetName())
}

func (exec *Executor) KillTask(exec.ExecutorDriver, *mesos.TaskID) {
	fmt.Println("Kill task")
}

func (exec *Executor) FrameworkMessage(driver exec.ExecutorDriver, msg string) {
	fmt.Println("Got framework message: ", msg)
}

func (exec *Executor) Shutdown(exec.ExecutorDriver) {
	fmt.Println("Shutting down the executor")
}

func (exec *Executor) Error(driver exec.ExecutorDriver, err string) {
	fmt.Println("Got error message:", err)
}

func init() {
	flag.Parse()
}

func main() {
	fmt.Println("Starting Example Executor (Go)")

	dconfig := exec.DriverConfig{
		Executor: newExampleExecutor(),
	}
	driver, err := exec.NewMesosExecutorDriver(dconfig)

	if err != nil {
		fmt.Println("Unable to create a ExecutorDriver ", err.Error())
	}

	_, err = driver.Start()
	if err != nil {
		fmt.Println("Got error:", err)
		return
	}
	fmt.Println("Executor process has started and running.")
	driver.Join()
}
