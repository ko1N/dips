package pipeline

import (
	"encoding/json"
	"strconv"

	log "github.com/inconshreveable/log15"
	"github.com/mattn/go-colorable"
)

// TODO: add timestamp for log/stdout/etc

// JobTracker - tracks job status, progress and logs
type JobTracker struct {
	log     log.Logger
	jobID   string
	taskID  uint
	channel chan string
}

// JobStatusType - the type of the job status update
type JobStatusType uint

const (
	// JobStatusLog - log info entry
	JobStatusLog JobStatusType = 0
	// JobStatusProgress - progress update
	JobStatusProgress JobStatusType = 1
	// JobStatusStdOut - stdout update
	JobStatusStdOut JobStatusType = 2
	// JobStatusStdErr - stderr update
	JobStatusStdErr JobStatusType = 3
)

// JobStatusMessage - a job status update message
type JobStatusMessage struct {
	JobID  string
	Type   JobStatusType
	TaskID uint
	Value  string
}

// CreateJobTracker - creates a new job tracking instance
func CreateJobTracker(rootlog log.Logger, channel chan string, jobID string) JobTracker {
	joblog := rootlog.New("job", jobID)
	joblog.SetHandler(log.MultiHandler(
		log.StreamHandler(colorable.NewColorableStdout(), log.TerminalFormat()),
		log.FuncHandler(func(r *log.Record) error {
			status, err := json.Marshal(JobStatusMessage{
				JobID:  jobID,
				Type:   JobStatusLog,
				TaskID: 0,
				Value:  r.Msg,
			})
			if err != nil {
				return err
			}
			channel <- string(status)
			return nil
		}),
	))

	tracker := JobTracker{
		log:     joblog,
		jobID:   jobID,
		taskID:  0,
		channel: channel,
	}
	tracker.Logger().Info("job `" + jobID + "` created")

	return tracker
}

// Logger - retruns the jobs logging instance
func (t *JobTracker) Logger() log.Logger {
	return t.log
}

// TrackTask - tracks a task change
func (t *JobTracker) TrackTask(taskID uint) {
	t.taskID = taskID
}

// TrackProgress - tracks progress of the current task
func (t *JobTracker) TrackProgress(progress uint) {
	//fmt.Printf("task %d progress: %d\n", t.taskID, progress)
	status, err := json.Marshal(JobStatusMessage{
		JobID:  t.jobID,
		Type:   JobStatusProgress,
		TaskID: t.taskID,
		Value:  strconv.Itoa(int(progress)), // TODO: fix types
	})
	if err != nil {
		//t.log.Crit("unable to marshal progress message")
		return
	}
	t.channel <- string(status)
}

// TrackStdOut - tracks stdout of the current task
func (t *JobTracker) TrackStdOut(outmsg string) {
	//fmt.Printf("task %d stdout: %s\n", t.taskID, outmsg)
	status, err := json.Marshal(JobStatusMessage{
		JobID:  t.jobID,
		Type:   JobStatusStdOut,
		TaskID: t.taskID,
		Value:  outmsg,
	})
	if err != nil {
		//t.log.Crit("unable to marshal stdout message")
		return
	}
	t.channel <- string(status)
}

// TrackStdErr - tracks stderr of the current task
func (t *JobTracker) TrackStdErr(errmsg string) {
	//fmt.Printf("task %d stderr: %s\n", t.taskID, errmsg)
	status, err := json.Marshal(JobStatusMessage{
		JobID:  t.jobID,
		Type:   JobStatusStdErr,
		TaskID: t.taskID,
		Value:  errmsg,
	})
	if err != nil {
		//t.log.Crit("unable to marshal stderr message")
		return
	}
	t.channel <- string(status)
}
