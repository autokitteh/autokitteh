-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;

-- Manually create a sequence column in current table
alter table session_log_records add column seq int;

-- Populate seq with the correct index in session log record
with cte as (
  select rowid, row_number() over (PARTITION by session_id order by rowid) as seq from  session_log_records
)

update session_log_records set seq = cte.seq from cte where session_log_records.rowid = cte.rowid;


-- create "new_session_log_records" table
CREATE TABLE `new_session_log_records` (
  `session_id` uuid NOT NULL,
  `seq` integer NULL NOT NULL,
  `data` json NULL,
  PRIMARY KEY (`session_id`, `seq`),
  CONSTRAINT `fk_session_log_records_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "session_log_records" to new temporary table "new_session_log_records"
INSERT INTO `new_session_log_records` (`session_id`, `data`, `seq`) SELECT `session_id`, `data`, `seq` FROM `session_log_records`;
-- drop "session_log_records" table after copying rows
DROP TABLE `session_log_records`;
-- rename temporary table "new_session_log_records" to "session_log_records"
ALTER TABLE `new_session_log_records` RENAME TO `session_log_records`;

-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_session_log_records" table
DROP TABLE `new_session_log_records`;
