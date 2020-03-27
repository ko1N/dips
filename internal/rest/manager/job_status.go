package manager

import (
	"encoding/json"
	"fmt"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

/*
func findTaskByID(job *model.Job, taskID uint) *model.JobStageTask {
	for _, s := range job.Stages {
		for _, t := range s.Tasks {
			if t.ID == taskID {
				return t
			}
		}
	}
	return nil
}

func updateJobProgress(job *model.Job) {
	var jobProgress float64
	for _, s := range job.Stages {
		var taskProgress float64
		for _, t := range s.Tasks {
			taskProgress += float64(t.Progress)
		}
		s.Progress = uint(taskProgress / float64(len(s.Tasks)))
		jobProgress += float64(s.Progress)
	}
	job.Progress = uint(jobProgress / float64(len(job.Stages)))
}
*/

func handleJobStatus() {
	for status := range recvJobStatus {
		msg := pipeline.JobStatusMessage{}
		err := json.Unmarshal([]byte(status), &msg)
		if err != nil {
			fmt.Printf("unable to unmarshal status message")
			continue
		}

		/*
			// find job
			job, err := jobs.FindJobByID(msg.JobID)
			if err != nil {
				fmt.Printf("unable to find job with id " + msg.JobID)
				continue
			}

			task := findTaskByID(job, msg.TaskID)
			if task == nil {
				fmt.Printf("unable to find task with id \"%d\"\n", msg.TaskID)
				continue
			}
		*/

		/*
			switch msg.Type {
			case pipeline.JobStatusLog:
				job.Logs = append(job.Logs, msg.Value)
				break

			case pipeline.JobStatusProgress:
				val, err := strconv.Atoi(msg.Value)
				if err != nil {
					fmt.Printf("unable to convert progress to int")
					continue
				}
				task.Progress = uint(val)
				updateJobProgress(job)
				break

			case pipeline.JobStatusStdOut:
				task.StdOut = append(task.StdOut, msg.Value)
				break

			case pipeline.JobStatusStdErr:
				task.StdErr = append(task.StdErr, msg.Value)
				break
			}

			job.Save()
		*/
	}
}
