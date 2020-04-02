package tracking

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
	// JobMessageStatus - stderr message
	JobMessageStatus JobMessageType = 0
	// JobMessageError - stderr message
	JobMessageError JobMessageType = 1
	// JobMessageStdIn - stdin message
	JobMessageStdIn JobMessageType = 2
	// JobMessageStdOut - stdout message
	JobMessageStdOut JobMessageType = 3
	// JobMessageStdErr - stderr message
	JobMessageStdErr JobMessageType = 4
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
			if conf.ProgressChannel == nil {
				return nil
			}

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

// TODO: DEPRECATED - RemoveMe
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
	if t.config.ProgressChannel == nil {
		return
	}

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

// Message - tracks messages of the current task
func (t *JobTracker) Message(mt JobMessageType, msg string) {
	if t.config.MessageChannel == nil || msg == "" {
		// do not persist empty messages
		return
	}

	//fmt.Printf("task %d stderr: %s\n", t.taskID, errmsg)
	status, err := json.Marshal(JobMessage{
		JobID:   t.config.JobID,
		Type:    mt,
		Message: msg,
	})
	if err != nil {
		//t.log.Crit("unable to marshal stderr message")
		return
	}
	t.config.MessageChannel <- string(status)
}

// Status - tracks a status message
func (t *JobTracker) Status(msg string) {
	t.Logger().Info(msg)
	t.Message(JobMessageStatus, msg)
}

func (t *JobTracker) Error(msg string, err error) {
	if err != nil {
		t.Logger().Crit(msg, "error", err)
		t.Message(JobMessageError, msg+" ("+err.Error()+")")
	} else {
		t.Logger().Crit(msg)
		t.Message(JobMessageError, msg)
	}
}

// StdIn - tracks a stdin message
func (t *JobTracker) StdIn(msg string) {
	t.Message(JobMessageStdIn, msg)
}

// StdOut - tracks a stdin message
func (t *JobTracker) StdOut(msg string) {
	t.Message(JobMessageStdOut, msg)
}

// StdErr - tracks a stdin message
func (t *JobTracker) StdErr(msg string) {
	t.Message(JobMessageStdErr, msg)
}
