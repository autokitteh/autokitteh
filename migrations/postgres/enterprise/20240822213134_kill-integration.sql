-- +goose Up
-- drop "integrations" table
DROP TABLE "integrations";

-- +goose Down
-- reverse: drop "integrations" table
CREATE TABLE "integrations" (
  "integration_id" uuid NOT NULL,
  "unique_name" text NULL,
  "display_name" text NULL,
  "description" text NULL,
  "logo_url" text NULL,
  "user_links" jsonb NULL,
  "connection_url" text NULL,
  "api_key" text NULL,
  "signing_key" text NULL,
  PRIMARY KEY ("integration_id")
);
CREATE UNIQUE INDEX "idx_integrations_unique_name" ON "integrations" ("unique_name");
