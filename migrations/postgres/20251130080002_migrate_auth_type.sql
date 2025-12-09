-- +goose Up
-- migrate "authType" variable name to "auth_type" in vars table
-- Insert auth_type copy for var_ids that only have authType (keep both for backward compatibility)
INSERT INTO "vars" ("var_id", "name", "value", "is_secret", "integration_id")
SELECT "var_id", 'auth_type', "value", "is_secret", "integration_id"
FROM "vars"
WHERE "name" = 'authType'
  AND NOT EXISTS (
    SELECT 1 FROM "vars" var
    WHERE var."name" = 'auth_type'
      AND var."var_id" = "vars"."var_id"
  );

-- +goose Down
-- reverse: migrate "authType" variable name to "auth_type" in vars table
-- Note: We cannot safely rollback this migration because we cannot distinguish between:
-- 1. auth_type entries that were added by this migration
-- 2. auth_type entries that already existed before the migration
-- Therefore, we do nothing on rollback to avoid data loss.
