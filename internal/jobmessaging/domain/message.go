package domain

import jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"

type Message struct {
	ID    MessageID
	JobID jobstatedomain.JobID
}
