package scheduler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"

	log "github.com/golang/glog"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
)

var (
	minCpu = 1.0
	minMem = 512.0
)

type Scheduler struct {
	executor      *mesos.ExecutorInfo
	tasksLaunched int
}

func NewLibvirtScheduler(exec *mesos.ExecutorInfo) (Scheduler, error) {
	return Scheduler{
		executor: exec,
	}, nil
}

func (sched *Scheduler) Registered(driver sched.SchedulerDriver, frameworkId *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Infoln("Scheduler Registered with Master ", masterInfo)
}

func (sched *Scheduler) Reregistered(driver sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Infoln("Scheduler Re-Registered with Master ", masterInfo)
}

func (sched *Scheduler) Disconnected(sched.SchedulerDriver) {
	log.Infoln("Scheduler Disconnected")
}

func (sched *Scheduler) OfferRescinded(s sched.SchedulerDriver, id *mesos.OfferID) {
	log.Infof("Offer '%v' rescinded.\n", *id)
}

func (sched *Scheduler) FrameworkMessage(s sched.SchedulerDriver, exId *mesos.ExecutorID, slvId *mesos.SlaveID, msg string) {
	log.Infof("Received framework message from executor '%v' on slave '%v': %s.\n", *exId, *slvId, msg)
}

func (sched *Scheduler) SlaveLost(s sched.SchedulerDriver, id *mesos.SlaveID) {
	log.Infof("Slave '%v' lost.\n", *id)
}

func (sched *Scheduler) ExecutorLost(s sched.SchedulerDriver, exId *mesos.ExecutorID, slvId *mesos.SlaveID, i int) {
	log.Infof("Executor '%v' lost on slave '%v' with exit code: %v.\n", *exId, *slvId, i)
}

func (sched *Scheduler) Error(driver sched.SchedulerDriver, err string) {
	log.Infoln("Scheduler received error:", err)
}

func (sched *Scheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log.Infof("Scheduler received resource offers:\n%s\n", strings.Join(offersStrings(offers), "\n"))

	for _, offer := range offers {
		// fuck it. dont even pull out the resources for this offer. just try to accept with some number of resources
		// and if we ask for more than the offer has, mesos should fail, right?

		var tasks []*mesos.TaskInfo
		if sched.tasksLaunched < 1 {

			log.Infof("Creating Task %d\n", sched.tasksLaunched)
			sched.tasksLaunched++

			taskId := &mesos.TaskID{
				Value: proto.String(strconv.Itoa(sched.tasksLaunched)),
			}

			task := &mesos.TaskInfo{
				Name:     proto.String("go-task-" + taskId.GetValue()),
				TaskId:   taskId,
				SlaveId:  offer.SlaveId,
				Executor: sched.executor,
				Resources: []*mesos.Resource{
					util.NewScalarResource("cpus", minCpu),
					util.NewScalarResource("mem", minMem),
				},
				Data: []byte("This is a payload string"),
			}
			log.Infof("Prepared task: %s with offer %s for launch\n", task.GetName(), offer.Id.GetValue())
			log.Infof("%+v\n", task)

			tasks = append(tasks, task)
		}
		if len(tasks) > 0 {
			log.Infoln("Launching", len(tasks), "tasks for offer", offer.Id.GetValue())
			driver.LaunchTasks([]*mesos.OfferID{offer.Id}, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(1)})
		}
	}

}

func (sched *Scheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Infoln("Status update: task", status.TaskId.GetValue(), " is in state ", status.State.Enum().String())
}

func offersStrings(offers []*mesos.Offer) []string {
	strs := make([]string, len(offers))
	for i, o := range offers {
		resources := make([]string, len(o.Resources))
		for j, r := range o.Resources {
			resources[j] = fmt.Sprintf("%s:%f", *r.Name, r.GetScalar().GetValue())
		}
		strs[i] = fmt.Sprintf("%s %s %s", *o.Id, *o.Hostname, strings.Join(resources, " "))
	}
	return strs
}
