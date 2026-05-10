package domain

type StaleJobRecoveryResult struct {
	Recovered     bool
	Terminal      bool
	AlreadyQueued bool
}
