CREATE TABLE IF NOT EXISTS execution (
    id SERIAL NOT NULL,
    type varchar(16) NOT NULL,
    status varchar(16),
    status_message text,
    trigger varchar(16) NOT NULL,
    extra_attrs JSONB,
    start_time timestamp DEFAULT CURRENT_TIMESTAMP,
    end_time timestamp NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS task (
    id SERIAL PRIMARY KEY NOT NULL,
    execution_id int NOT NULL,
    job_id varchar(64),
    status varchar(16) NOT NULL,
    status_code int NOT NULL,
    status_revision int,
    status_message text,
    retry_count int,
    extra_attrs JSONB,
    start_time timestamp DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp DEFAULT CURRENT_TIMESTAMP,
    end_time timestamp NULL,
    FOREIGN KEY (execution_id) REFERENCES execution(id)
);

CREATE TABLE IF NOT EXISTS check_in_data (
    id SERIAL PRIMARY KEY NOT NULL,
    task_id int NOT NULL,
    data text,
    creation_time timestamp DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES task(id)
);