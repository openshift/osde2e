CREATE TYPE job_result AS ENUM ('passed', 'failed', 'aborted');

CREATE TABLE IF NOT EXISTS jobs (
    id bigserial PRIMARY KEY,
    provider text NOT NULL,
    job_name text NOT NULL,
    job_id text NOT NULL,
    url text NOT NULL,
    started timestamp with time zone NOT NULL,
    finished timestamp with time zone NOT NULL,
    duration interval GENERATED ALWAYS AS (finished - started) STORED,
    cluster_version text NOT NULL,
    cluster_name text NOT NULL,
    cluster_id text NOT NULL,
    multi_az text NOT NULL,
    channel text NOT NULL,
    environment text NOT NULL,
    region text NOT NULL,
    numb_worker_nodes integer NOT NULL,
    network_provider text NOT NULL,
    image_content_source text NOT NULL,
    install_config text NOT NULL,
    hibernate_after_use boolean NOT NULL,
    reused boolean NOT NULL,
    result job_result NOT NULL
);

