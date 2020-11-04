ALTER TABLE schedule ADD COLUMN IF NOT EXISTS cron_type varchar(64);

/*fill the task count, status and end_time of execution based on the tasks*/
DO $$
DECLARE
    rep_exec RECORD;
    status_count RECORD;
    rep_status varchar(32);
BEGIN
    FOR rep_exec IN SELECT * FROM replication_execution
    LOOP
      /*the replication status is set directly in some cases, so skip if the status is a final one*/
      IF rep_exec.status='Stopped' OR rep_exec.status='Failed' OR rep_exec.status='Succeed' THEN
        CONTINUE;
      END IF;
      /*fulfill the status count*/
      FOR status_count IN SELECT status, COUNT(*) as c FROM replication_task WHERE execution_id=rep_exec.id GROUP BY status
      LOOP
        IF status_count.status = 'Stopped' THEN
          UPDATE replication_execution SET stopped=status_count.c WHERE id=rep_exec.id;
        ELSIF status_count.status = 'Failed' THEN
          UPDATE replication_execution SET failed=status_count.c WHERE id=rep_exec.id;
        ELSIF status_count.status = 'Succeed' THEN
          UPDATE replication_execution SET succeed=status_count.c WHERE id=rep_exec.id;
        ELSE
          UPDATE replication_execution SET in_progress=status_count.c WHERE id=rep_exec.id;
        END IF;
      END LOOP;

      /*reload the execution record*/
      SELECT * INTO rep_exec FROM replication_execution where id=rep_exec.id;

      /*calculate the status*/
      IF rep_exec.in_progress>0 THEN
        rep_status = 'InProgress';
      ELSIF rep_exec.failed>0 THEN
        rep_status = 'Failed';
      ELSIF rep_exec.stopped>0 THEN
        rep_status = 'Stopped';
      ELSE
        rep_status = 'Succeed';
      END IF;
      UPDATE replication_execution SET status=rep_status WHERE id=rep_exec.id;

      /*update the end time if the status is a final one*/
      IF rep_status='Failed' OR rep_status='Stopped' OR rep_status='Succeed' THEN
        UPDATE replication_execution
            SET end_time=(SELECT MAX (end_time) FROM replication_task WHERE execution_id=rep_exec.id)
            WHERE id=rep_exec.id;
      END IF;
    END LOOP;
END $$;

/*move the replication execution records into the new execution table*/
ALTER TABLE replication_execution ADD COLUMN IF NOT EXISTS new_execution_id int;
DO $$
DECLARE
    rep_exec RECORD;
    trigger varchar(64);
    status varchar(32);
    new_exec_id integer;
BEGIN
    FOR rep_exec IN SELECT * FROM replication_execution
    LOOP
      IF rep_exec.trigger = 'scheduled' THEN
        trigger = 'SCHEDULE';
      ELSIF rep_exec.trigger = 'event_based' THEN
        trigger = 'EVENT';
      ELSE
        trigger = 'MANUAL';
      END IF;

      IF rep_exec.status = 'InProgress' THEN
        status = 'Running';
      ELSIF rep_exec.status = 'Stopped' THEN
        status = 'Stopped';
      ELSIF rep_exec.status = 'Failed' THEN
        status = 'Error';
      ELSIF rep_exec.status = 'Succeed' THEN
        status = 'Success';
      END IF;
      
      INSERT INTO execution (vendor_type, vendor_id, status, status_message, trigger, start_time, end_time)
        VALUES ('REPLICATION', rep_exec.policy_id, status, rep_exec.status_text, trigger, rep_exec.start_time, rep_exec.end_time) RETURNING id INTO new_exec_id;
      UPDATE replication_execution SET new_execution_id=new_exec_id WHERE id=rep_exec.id;
    END LOOP;
END $$;

/*move the replication task records into the new task table*/
DO $$
DECLARE
    rep_task RECORD;
    status varchar(32);
    status_code integer;
BEGIN
    FOR rep_task IN SELECT * FROM replication_task
    LOOP
      IF rep_task.status = 'InProgress' THEN
        status = 'Running';
        status_code = 2;
      ELSIF rep_task.status = 'Stopped' THEN
        status = 'Stopped';
        status_code = 3;
      ELSIF rep_task.status = 'Failed' THEN
        status = 'Error';
        status_code = 3;
      ELSIF rep_task.status = 'Succeed' THEN
        status = 'Success';
        status_code = 3;
      ELSE
        status = 'Pending';
        status_code = 0;
      END IF;
      INSERT INTO task (execution_id, job_id, status, status_code, status_revision,
        run_count, extra_attrs, creation_time, start_time, update_time, end_time)
        VALUES ((SELECT new_execution_id FROM replication_execution WHERE id=rep_task.execution_id),
            rep_task.job_id, status, status_code, rep_task.status_revision,
            1, CONCAT('{"resource_type":"', rep_task.resource_type,'","source_resource":"', rep_task.src_resource, '","destination_resource":"', rep_task.dst_resource, '","operation":"', rep_task.operation,'"}')::json,
            rep_task.start_time, rep_task.start_time, rep_task.end_time, rep_task.end_time);
      END LOOP;
END $$;

DROP TABLE IF EXISTS replication_task;
DROP TABLE IF EXISTS replication_execution;