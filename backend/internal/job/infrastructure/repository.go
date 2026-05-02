package infrastructure

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (repo *JobRepository) Create(ctx context.Context, job *domain.Job) error {
	const query = `
		INSERT INTO jobs (
			job_id,
			model,
			messages,
			generation_params,
			status,
			final_result,
			error_message,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := repo.db.ExecContext(
		ctx,
		query,
		job.JobID,
		job.Model,
		job.Messages,
		job.GenerationParams,
		job.Status,
		job.FinalResult,
		job.ErrorMessage,
		job.CreatedAt,
		job.UpdatedAt,
	)

	return err
}

func (repo *JobRepository) GetByID(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	const query = `
		SELECT
			job_id,
			model,
			messages,
			generation_params,
			status,
			final_result,
			error_message,
			created_at,
			updated_at,
		FROM jobs
		WHERE job_id = $1
	`

	job := &domain.Job{}

	err := repo.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.JobID,
		&job.Model,
		&job.Messages,
		&job.GenerationParams,
		&job.Status,
		&job.FinalResult,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (repo *JobRepository) Update(ctx context.Context, job *domain.Job) error {
	const query = `
		UPDATE jobs
		SET
			model = $2,
			messages = $3,
			generation_params = $4,
			status = $5,
			final_result = $6,
			error_message = $7,
			updated_at = $8
		WHERE job_id = $1
	`
	_, err := repo.db.ExecContext(
		ctx,
		query,
		job.JobID,
		job.Model,
		job.Messages,
		job.GenerationParams,
		job.Status,
		job.FinalResult,
		job.ErrorMessage,
		job.UpdatedAt,
	)

	return err
}
