-- +goose Up
CREATE TABLE applications(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    cover_note TEXT NOT NULL,
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    CONSTRAINT job_user_constraint UNIQUE (job_id, user_id)
);

-- +goose Down
DROP TABLE applications;