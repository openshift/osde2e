-- name: CreateJob :one
INSERT INTO jobs (
    provider,
    job_name,
    job_id,
    url,
    started,
    finished,
    cluster_version,
    upgrade_version,
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
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
    job_id,
    result,
    name,
    duration,
    error,
    stdout,
    stderr
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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

-- name: ListAlertableFailuresForJob :many
select 
    jobs.*,
    -- remove the job phase from the test name
    regexp_replace(testcases.name, '\[(install|upgrade)\] (.*)', '\2') as name,
    testcases.result as testresult
from jobs
    join testcases
    on jobs.id = testcases.job_id
where
    jobs.id = sqlc.arg(jobID)
    -- filter kinds of test we do not care about
    and testcases.name !~ '.*\[Suite: (informing|addons|conformance)\].*'
    -- ensure this test does belong to a suite
    and testcases.name ~ '.*\[Suite:.*'
    and testcases.result != 'passed'
    and testcases.result != 'skipped'
;

-- name: ListAlertableRecentTestFailures :many
with testcases as (
    select 
        jobs.*,
        -- remove the job phase from the test name
        regexp_replace(testcases.name, '\[(install|upgrade)\] (.*)', '\2') as name,
        testcases.result as testresult
    from jobs
        join testcases
        on jobs.id = testcases.job_id
    where
        now() - jobs.started < interval '48 hours'
        and (testcases.result = 'failure' or testcases.result = 'error')
)
select *
from testcases
where
    testcases.name = ANY(sqlc.arg(names)::text[])
;

-- name: ListProblematicTests :many
with recent_tests as (
    select 
        jobs.*,
        regexp_replace(name, '\[(install|upgrade)\] (.*)', '\2') as name,
        testcases.result as testresult
    from jobs
        join testcases
        on jobs.id = testcases.job_id
    where
        -- filter kinds of test we do not care about
        testcases.name !~ '.*\[Suite: (informing|addons|conformance)\].*'
        -- ensure this test does belong to a suite
        and testcases.name ~ '.*\[Suite:.*'
        and testcases.name !~ '.*sig-.*'
        and now() - jobs.started < interval '48 hours'
        -- filter out osde2e's own CI jobs
        and jobs.job_id != '-1'
), counts as (
        -- synthesize a table with the name of a test and columns counting how often it has resulted
        -- in each result type
        select
            name,
            count(CASE WHEN recent_tests.testresult='failure' THEN 1 END) as failure,
            count(CASE WHEN recent_tests.testresult='error' THEN 1 END) as error
        from recent_tests
        group by name
)
select 
    counts.name,
    (counts.error + counts.failure) as problems
from
    counts
where counts.error + counts.failure > 1
;

