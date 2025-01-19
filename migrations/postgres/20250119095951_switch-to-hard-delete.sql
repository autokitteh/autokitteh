-- +goose Up
-- modify "projects" table
ALTER TABLE "projects" DROP COLUMN "deleted_at";
-- modify "builds" table
ALTER TABLE "builds" DROP CONSTRAINT "fk_builds_project", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_builds_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "connections" table
ALTER TABLE "connections" DROP CONSTRAINT "fk_connections_project", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_connections_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "deployments" table
ALTER TABLE "deployments" DROP CONSTRAINT "fk_deployments_build", DROP CONSTRAINT "fk_deployments_project", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_deployments_build" FOREIGN KEY ("build_id") REFERENCES "builds" ("build_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE, ADD
 CONSTRAINT "fk_deployments_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "triggers" table
ALTER TABLE "triggers" DROP CONSTRAINT "fk_triggers_connection", DROP CONSTRAINT "fk_triggers_project", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_triggers_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE, ADD
 CONSTRAINT "fk_triggers_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_connection", DROP CONSTRAINT "fk_events_project", DROP CONSTRAINT "fk_events_trigger", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_events_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
 CONSTRAINT "fk_events_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE, ADD
 CONSTRAINT "fk_events_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_build", DROP CONSTRAINT "fk_sessions_deployment", DROP CONSTRAINT "fk_sessions_event", DROP CONSTRAINT "fk_sessions_project", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_sessions_build" FOREIGN KEY ("build_id") REFERENCES "builds" ("build_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE, ADD
 CONSTRAINT "fk_sessions_deployment" FOREIGN KEY ("deployment_id") REFERENCES "deployments" ("deployment_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE, ADD
 CONSTRAINT "fk_sessions_event" FOREIGN KEY ("event_id") REFERENCES "events" ("event_id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
 CONSTRAINT "fk_sessions_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "session_call_attempts" table
ALTER TABLE "session_call_attempts" DROP CONSTRAINT "fk_session_call_attempts_session", ADD
 CONSTRAINT "fk_session_call_attempts_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "session_call_specs" table
ALTER TABLE "session_call_specs" DROP CONSTRAINT "fk_session_call_specs_session", ADD
 CONSTRAINT "fk_session_call_specs_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "session_log_records" table
ALTER TABLE "session_log_records" DROP CONSTRAINT "fk_session_log_records_session", ADD
 CONSTRAINT "fk_session_log_records_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "signals" table
ALTER TABLE "signals" DROP CONSTRAINT "fk_signals_connection", DROP CONSTRAINT "fk_signals_trigger", ADD
 CONSTRAINT "fk_signals_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE, ADD
 CONSTRAINT "fk_signals_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;
-- modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "fk_values_project", ADD
 CONSTRAINT "fk_values_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE;

-- +goose Down
-- reverse: modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "fk_values_project", ADD
 CONSTRAINT "fk_values_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "signals" table
ALTER TABLE "signals" DROP CONSTRAINT "fk_signals_trigger", DROP CONSTRAINT "fk_signals_connection", ADD
 CONSTRAINT "fk_signals_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_signals_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "session_log_records" table
ALTER TABLE "session_log_records" DROP CONSTRAINT "fk_session_log_records_session", ADD
 CONSTRAINT "fk_session_log_records_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "session_call_specs" table
ALTER TABLE "session_call_specs" DROP CONSTRAINT "fk_session_call_specs_session", ADD
 CONSTRAINT "fk_session_call_specs_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "session_call_attempts" table
ALTER TABLE "session_call_attempts" DROP CONSTRAINT "fk_session_call_attempts_session", ADD
 CONSTRAINT "fk_session_call_attempts_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_project", DROP CONSTRAINT "fk_sessions_event", DROP CONSTRAINT "fk_sessions_deployment", DROP CONSTRAINT "fk_sessions_build", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_sessions_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_sessions_event" FOREIGN KEY ("event_id") REFERENCES "events" ("event_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_sessions_deployment" FOREIGN KEY ("deployment_id") REFERENCES "deployments" ("deployment_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_sessions_build" FOREIGN KEY ("build_id") REFERENCES "builds" ("build_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_trigger", DROP CONSTRAINT "fk_events_project", DROP CONSTRAINT "fk_events_connection", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_events_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_events_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_events_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP CONSTRAINT "fk_triggers_project", DROP CONSTRAINT "fk_triggers_connection", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_triggers_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_triggers_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "deployments" table
ALTER TABLE "deployments" DROP CONSTRAINT "fk_deployments_project", DROP CONSTRAINT "fk_deployments_build", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_deployments_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_deployments_build" FOREIGN KEY ("build_id") REFERENCES "builds" ("build_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "connections" table
ALTER TABLE "connections" DROP CONSTRAINT "fk_connections_project", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_connections_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "builds" table
ALTER TABLE "builds" DROP CONSTRAINT "fk_builds_project", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_builds_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "projects" table
ALTER TABLE "projects" ADD COLUMN "deleted_at" timestamptz NULL;
