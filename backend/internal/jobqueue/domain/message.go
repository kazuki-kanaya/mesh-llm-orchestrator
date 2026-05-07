package domain

import jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"

type Message struct {
	ID    MessageID
	JobID jobstatedomain.JobID
}
