package domain

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

func (s Status) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

func (s Status) IsRunnable() bool {
	return s == StatusQueued
}

func (s Status) IsValid() bool {
	switch s {
	case StatusQueued, StatusRunning, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}
