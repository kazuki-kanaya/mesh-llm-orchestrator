package redis

import jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"

func JobKey(jobID jobstatedomain.JobID) string {
	return "job:" + jobID.String()
}

func JobResultChannel(jobID jobstatedomain.JobID) string {
	return "result:" + jobID.String()
}

func JobStreamKey() string {
	return "stream:jobs"
}

func RunningJobsKey() string {
	return "jobs:running"
}
