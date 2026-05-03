package redis

import "github.com/google/uuid"

func JobKey(jobID uuid.UUID) string {
	return "job:" + jobID.String()
}

func JobResultChannel(jobID uuid.UUID) string {
	return "result:" + jobID.String()
}

func JobQueueKey() string {
	return "queue:jobs"
}
