package redis

func JobKey(jobID string) string {
	return "job:" + jobID
}

func JobResultChannel(jobID string) string {
	return "result:" + jobID
}

func JobStreamKey() string {
	return "stream:jobs"
}

func JobConsumerGroupName() string {
	return "job-executors"
}

func RunningJobsKey() string {
	return "jobs:running"
}
