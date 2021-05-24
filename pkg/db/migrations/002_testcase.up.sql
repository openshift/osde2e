CREATE TYPE test_result as ENUM ('passed', 'failure', 'skipped', 'error');

CREATE TABLE IF NOT EXISTS testcases (
    id bigserial PRIMARY KEY,
    job_id bigserial REFERENCES jobs NOT NULL,
    result test_result NOT NULL,
    name text NOT NULL,
    duration interval NOT NULL,
    error text NOT NULL,
    stdout text NOT NULL,
    stderr text NOT NULL
);
