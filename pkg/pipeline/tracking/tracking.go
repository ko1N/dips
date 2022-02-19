package tracking

import (
	log "github.com/inconshreveable/log15"
	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/mattn/go-colorable"
)

// TODO: add timestamp for log/stdout/etc

// JobTrackerConfig - config for a job tracker instance
type JobTrackerConfig struct {
	Logger log.Logger
	Client *dipscl.Client
	JobId  string
}

// JobTracker - tracks job status, progress and logs
type JobTracker struct {
	config  JobTrackerConfig
	jobLog  log.Logger
	taskIdx uint
}

// CreateJobTracker - creates a new job tracking instance
func CreateJobTracker(conf *JobTrackerConfig) JobTracker {
	jobLog := conf.Logger.New("job", conf.JobId)
	jobLog.SetHandler(log.MultiHandler(
		log.StreamHandler(colorable.NewColorableStdout(), log.TerminalFormat()),
		log.FuncHandler(func(r *log.Record) error {
			if conf.Client == nil {
				return nil
			}
			conf.Client.NewEvent().
				Message(&dipscl.MessageEvent{
					JobId:   conf.JobId,
					TaskIdx: 0,
					Type:    dipscl.StatusMessage,
					Message: r.Msg,
				}).
				Dispatch()
			return nil
		}),
	))

	tracker := JobTracker{
		config:  *conf,
		jobLog:  jobLog,
		taskIdx: 0,
	}
	tracker.Logger().Info("tracker for job `" + conf.JobId + "` created")

	return tracker
}

// TODO: DEPRECATED - RemoveMe
// Logger - retruns the jobs logging instance
func (t *JobTracker) Logger() log.Logger {
	return t.jobLog
}

// TrackTask - tracks a task change
func (t *JobTracker) TrackTask(taskIdx uint) {
	t.taskIdx = taskIdx
}

// TrackProgress - tracks progress of the current task
func (t *JobTracker) TrackProgress(progress uint) {
	if t.config.Client == nil {
		return
	}
	t.config.Client.NewEvent().
		Status(&dipscl.StatusEvent{
			JobId:    t.config.JobId,
			TaskIdx:  t.taskIdx,
			Type:     dipscl.ProgressEvent,
			Progress: progress,
		}).
		Dispatch()
}

// Message - tracks messages of the current task
func (t *JobTracker) Message(mt dipscl.MessageEventType, msg string) {
	if t.config.Client == nil || msg == "" {
		// do not persist empty messages
		return
	}

	//fmt.Printf("task %d stderr: %s\n", t.taskID, errmsg)
	t.config.Client.NewEvent().
		Message(&dipscl.MessageEvent{
			JobId:   t.config.JobId,
			TaskIdx: t.taskIdx,
			Type:    mt,
			Message: msg,
		}).
		Dispatch()
}

// Status - tracks a status message
func (t *JobTracker) Status(msg string) {
	t.Logger().Info(msg)
	t.Message(dipscl.StatusMessage, msg)
}

func (t *JobTracker) Error(msg string, err error) {
	if err != nil {
		t.Logger().Crit(msg, "error", err)
		t.Message(dipscl.ErrorMessage, msg+" ("+err.Error()+")")
	} else {
		t.Logger().Crit(msg)
		t.Message(dipscl.ErrorMessage, msg)
	}
}

// StdIn - tracks a stdin message
func (t *JobTracker) StdIn(msg string) {
	t.Message(dipscl.StdInMessage, msg)
}

// StdOut - tracks a stdin message
func (t *JobTracker) StdOut(msg string) {
	t.Message(dipscl.StdOutMessage, msg)
}

// StdErr - tracks a stdin message
func (t *JobTracker) StdErr(msg string) {
	t.Message(dipscl.StdErrMessage, msg)
}
