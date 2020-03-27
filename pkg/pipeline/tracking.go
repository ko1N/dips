package pipeline

import (
	"encoding/json"
	"strconv"

	log "github.com/inconshreveable/log15"
	"github.com/mattn/go-colorable"
)

// TODO: add timestamp for log/stdout/etc

// JobTrackerConfig - config for a job tracker instance
type JobTrackerConfig struct {
	Logger          log.Logger
	ProgressChannel chan string
	MessageChannel  chan string
	JobID           string
}

// JobTracker - tracks job status, progress and logs
type JobTracker struct {
	config JobTrackerConfig
	jobLog log.Logger
	taskID uint
}

// TODO: this should be in a seperate amqp model ->

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

// JobMessageType - the type of the job message
type JobMessageType uint

const (
	// JobMessageStdOut - stdout message
	JobMessageStdOut JobMessageType = 0
	// JobMessageStdErr - stderr message
	JobMessageStdErr JobMessageType = 1
)

// JobMessage - describes a jobs message
type JobMessage struct {
	JobID   string
	Type    JobMessageType
	Message string
}

// CreateJobTracker - creates a new job tracking instance
func CreateJobTracker(conf JobTrackerConfig) JobTracker {
	jobLog := conf.Logger.New("job", conf.JobID)
	jobLog.SetHandler(log.MultiHandler(
		log.StreamHandler(colorable.NewColorableStdout(), log.TerminalFormat()),
		log.FuncHandler(func(r *log.Record) error {
			status, err := json.Marshal(JobStatusMessage{
				JobID:  conf.JobID,
				Type:   JobStatusLog,
				TaskID: 0,
				Value:  r.Msg,
			})
			if err != nil {
				return err
			}
			conf.ProgressChannel <- string(status)
			return nil
		}),
	))

	tracker := JobTracker{
		config: conf,
		jobLog: jobLog,
		taskID: 0,
	}
	tracker.Logger().Info("job `" + conf.JobID + "` created")

	return tracker
}

// Logger - retruns the jobs logging instance
func (t *JobTracker) Logger() log.Logger {
	return t.jobLog
}

// TrackTask - tracks a task change
func (t *JobTracker) TrackTask(taskID uint) {
	t.taskID = taskID
}

// TrackProgress - tracks progress of the current task
func (t *JobTracker) TrackProgress(progress uint) {
	//fmt.Printf("task %d progress: %d\n", t.taskID, progress)
	status, err := json.Marshal(JobStatusMessage{
		JobID:  t.config.JobID,
		Type:   JobStatusProgress,
		TaskID: t.taskID,
		Value:  strconv.Itoa(int(progress)), // TODO: fix types
	})
	if err != nil {
		//t.log.Crit("unable to marshal progress message")
		return
	}
	t.config.ProgressChannel <- string(status)
}

// TrackStdOut - tracks stdout of the current task
func (t *JobTracker) TrackStdOut(outmsg string) {
	//fmt.Printf("task %d stderr: %s\n", t.taskID, errmsg)
	status, err := json.Marshal(JobMessage{
		JobID:   t.config.JobID,
		Type:    JobMessageStdOut,
		Message: outmsg,
	})
	if err != nil {
		//t.log.Crit("unable to marshal stderr message")
		return
	}
	t.config.MessageChannel <- string(status)
}

// TrackStdErr - tracks stderr of the current task
func (t *JobTracker) TrackStdErr(errmsg string) {
	//fmt.Printf("task %d stderr: %s\n", t.taskID, errmsg)
	status, err := json.Marshal(JobMessage{
		JobID:   t.config.JobID,
		Type:    JobMessageStdErr,
		Message: errmsg,
	})
	if err != nil {
		//t.log.Crit("unable to marshal stderr message")
		return
	}
	t.config.MessageChannel <- string(status)
}
