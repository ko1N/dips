package tracking

import (
	"fmt"

	log "github.com/inconshreveable/log15"
	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/mattn/go-colorable"
)

// Tracks job status, progress and logs
type JobTracker struct {
	logger log.Logger
	client *dipscl.Client
	jobId  string
	taskId string
}

// Creates a new job tracking instance
func CreateJobTracker(logger log.Logger, cl *dipscl.Client, jobId string) JobTracker {
	l := logger.New("job", jobId)
	l.SetHandler(log.MultiHandler(
		log.StreamHandler(colorable.NewColorableStdout(), log.TerminalFormat()),
	))

	tracker := JobTracker{
		logger: l,
		client: cl,
		jobId:  jobId,
		taskId: "",
	}
	tracker.Info("tracker for job `" + jobId + "` created")
	return tracker
}

// Creates a new task tracking instance
func CreateTaskTracker(logger log.Logger, cl *dipscl.Client, jobId string, taskId string) JobTracker {
	l := logger.New("job", jobId, "task", taskId)
	l.SetHandler(log.MultiHandler(
		log.StreamHandler(colorable.NewColorableStdout(), log.TerminalFormat()),
	))

	tracker := JobTracker{
		logger: l,
		client: cl,
		jobId:  jobId,
		taskId: taskId,
	}
	tracker.Info("tracker for task `" + taskId + "` created")
	return tracker
}

// Tracks progress of the current task
func (t *JobTracker) Progress(progress uint) {
	if t.client == nil {
		return
	}
	t.client.NewEvent().
		Status(&dipscl.StatusEvent{
			JobId:    t.jobId,
			TaskId:   t.taskId,
			Type:     dipscl.ProgressEvent,
			Progress: progress,
		}).
		Dispatch()
}

func (t *JobTracker) log(ty dipscl.MessageEventType, msg string) {
	if t.client == nil || msg == "" {
		// do not persist empty messages
		return
	}

	t.client.NewEvent().
		Message(&dipscl.MessageEvent{
			JobId:   t.jobId,
			TaskId:  t.taskId,
			Type:    ty,
			Message: msg,
		}).
		Dispatch()
}

func (t *JobTracker) Debug(msg string, ctx ...interface{}) {
	t.logger.Debug(msg, ctx...)
	t.log(dipscl.LogDebugMessage, fmt.Sprintf(msg, ctx...))
}

func (t *JobTracker) Info(msg string, ctx ...interface{}) {
	t.logger.Info(msg, ctx...)
	t.log(dipscl.LogInfoMessage, fmt.Sprintf(msg, ctx...))
}

func (t *JobTracker) Warn(msg string, ctx ...interface{}) {
	t.logger.Warn(msg, ctx...)
	t.log(dipscl.LogWarnMessage, fmt.Sprintf(msg, ctx...))
}

func (t *JobTracker) Error(msg string, ctx ...interface{}) {
	t.logger.Error(msg, ctx...)
	t.log(dipscl.LogErrorMessage, fmt.Sprintf(msg, ctx...))
}

func (t *JobTracker) Crit(msg string, ctx ...interface{}) {
	t.logger.Crit(msg, ctx...)
	t.log(dipscl.LogCritMessage, fmt.Sprintf(msg, ctx...))
}

func (t *JobTracker) StdOut(msg string, ctx ...interface{}) {
	fmt.Println(msg)
	t.log(dipscl.StdOutMessage, fmt.Sprintf(msg, ctx...))
}

func (t *JobTracker) StdErr(msg string, ctx ...interface{}) {
	fmt.Println(msg)
	t.log(dipscl.StdErrMessage, fmt.Sprintf(msg, ctx...))
}
