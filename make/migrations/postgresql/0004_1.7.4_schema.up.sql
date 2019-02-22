/*delete replication jobs whose policy has been marked as "deleted"*/
DELETE FROM replication_job AS j
USING replication_policy AS p
WHERE j.policy_id = p.id AND p.deleted = TRUE;

/*delete replication policy which has been marked as "deleted"*/
DELETE FROM replication_policy AS p
WHERE p.deleted = TRUE; 