-- name: CreateJob :one
INSERT INTO jobs (
    provider,
    job_name,
    job_id,
    url,
    started,
    finished,
    cluster_version,
    cluster_name,
    cluster_id,
    multi_az,
    channel,
    environment,
    region,
    numb_worker_nodes,
    network_provider,
    image_content_source,
    install_config,
    hibernate_after_use,
    reused,
    result
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
RETURNING id;

-- name: GetJob :one
SELECT *
FROM jobs
WHERE jobs.id = $1;

-- name: ListJobs :many
SELECT *
FROM jobs
ORDER BY id;

-- name: CreateTestcase :one
INSERT INTO testcases (
    id,
    job_id,
    result,
    name,
    duration,
    error,
    stdout,
    stderr
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: GetTestcase :one
SELECT *
FROM testcases
WHERE testcases.id = $1;

-- name: GetTestcaseForJob :many
SELECT *
from testcases
WHERE testcases.job_id = $1
ORDER BY testcases.id;

-- name: ListTestcases :many
SELECT *
FROM testcases
ORDER BY id;

