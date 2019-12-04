/* TODO remove the table artifact_2 and use the artifact instead after finishing the upgrade work */
CREATE TABLE artifact_2
(
  id            SERIAL PRIMARY KEY NOT NULL,
  /* image, chart, etc */
  type          varchar(255),
  project_id    int NOT NULL,
  repository_id int NOT NULL,
  media_type    varchar(255),
  digest        varchar(255) NOT NULL,
  size          bigint,  
  upload_time   timestamp default CURRENT_TIMESTAMP,
  extra_attrs   text,
  annotations   text,
  CONSTRAINT unique_artifact_2 UNIQUE (repository_id, digest)
);

CREATE TABLE tag
(
  id            SERIAL PRIMARY KEY NOT NULL,
  repository_id int NOT NULL,
  artifact_id   int NOT NULL,
  name          varchar(255),
  upload_time   timestamp default CURRENT_TIMESTAMP,
  latest_download_time timestamp,
  CONSTRAINT unique_tag UNIQUE (repository_id, name)
);

/* artifact_reference records the child artifact referenced by parent artifact */
CREATE TABLE artifact_reference
(
  id            SERIAL PRIMARY KEY NOT NULL,
  artifact_id   int NOT NULL,
  reference_id  int NOT NULL,
  platform      varchar(255),
  CONSTRAINT unique_reference UNIQUE (artifact_id, reference_id)
);