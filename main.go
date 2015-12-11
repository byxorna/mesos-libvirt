package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	byxscheduler "github.com/byxorna/mesos-libvirt/scheduler"
	"github.com/gogo/protobuf/proto"

	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
)

var (
	// filled in by makefile
	version = "development"
	commit  = "?"
	branch  = "?"

	address = "0.0.0.0"
	//address      = flag.String("address", "127.0.0.1", "Binding address for artifact server")
	artifactPort = flag.Int("artifactPort", 8080, "Binding port for artifact server")
	master       = flag.String("master", "127.0.0.1:5050", "Master address <ip:port>")
	//executorPath = flag.String("executor", "executor", "Path to executor on HTTP")
)

func init() {
	flag.Parse()
}

func main() {

	// Start HTTP server hosting executor binary
	go func() {
		x := fmt.Sprintf("%s:%d", address, *artifactPort)
		log.Println("Starting up artifact server on " + x)
		http.ListenAndServe(x, http.FileServer(http.Dir(".")))
	}()
	//uri := ServeExecutorArtifact(*address, *artifactPort, *executorPath)
	executorArtifactUri := fmt.Sprintf("http://framework:%d/executor/executor", *artifactPort)

	// Executor
	exec := prepareExecutorInfo(executorArtifactUri, "./executor")

	// Scheduler
	scheduler, err := byxscheduler.NewLibvirtScheduler(exec)
	if err != nil {
		log.Fatalf("Failed to create scheduler with error: %v\n", err)
		os.Exit(-2)
	}

	// Framework
	fwinfo := &mesos.FrameworkInfo{
		User: proto.String(""), // Mesos-go will fill in user.
		Name: proto.String("Libvirt Framework (" + version + ")"),
	}

	// Scheduler Driver
	config := sched.DriverConfig{
		Scheduler:      &scheduler,
		Framework:      fwinfo,
		Master:         *master,
		Credential:     (*mesos.Credential)(nil),
		BindingAddress: net.ParseIP(address),
	}

	driver, err := sched.NewMesosSchedulerDriver(config)

	if err != nil {
		log.Fatalf("Unable to create a SchedulerDriver: %v\n", err.Error())
		os.Exit(-3)
	}

	if stat, err := driver.Run(); err != nil {
		log.Fatalf("Framework stopped with status %s and error: %s\n", stat.String(), err.Error())
		os.Exit(-4)
	}
}

func prepareExecutorInfo(uri string, cmd string) *mesos.ExecutorInfo {
	executorUris := []*mesos.CommandInfo_URI{
		{
			Value:      &uri,
			Executable: proto.Bool(true),
		},
	}

	return &mesos.ExecutorInfo{
		ExecutorId: util.NewExecutorID("default"),
		Name:       proto.String("Libvirt Executor (" + version + ")"),
		Source:     proto.String("go_test"),
		Command: &mesos.CommandInfo{
			Value: proto.String(cmd),
			Uris:  executorUris,
		},
	}
}
