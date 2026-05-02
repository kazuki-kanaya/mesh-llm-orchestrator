CREATE TABLE jobs (
    job_id UUID PRIMARY KEY,
    model TEXT NOT NULL,
    messages JSONB NOT NULL,
    generation_params JSONB NOT NULL,
    status TEXT NOT NULL,
    final_result TEXT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT jobs_status_check CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled'))
);

CREATE INDEX idx_jobs_status ON jobs (status);
CREATE INDEX idx_jobs_created_at ON jobs (created_at);