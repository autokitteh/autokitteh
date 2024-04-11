# Database Migrations

We use [Atlas](http://atlasgo.io/) to generate migrations based on GORM's models.

## How to generate a new migration
- Update the DB models that you need
- Run `make generate-migrations`, give a meaningful name to the migration (i.e. `add-created-at-to-projects`)

### Resolving conflicts in migrations
In case there are migrations created on two different branches there would be a conflict which should be resolved

First branch merging to main would work fine. Second branch would have to rebase from main and run ```atlas migrate rebase <versions>``` from the problematic versions
. Consule [Atlas Docs](https://atlasgo.io/versioned/apply) for further explanations.

## Running migrations
if you set `AK_DB__AUTO_MIGRATE=true` then migrations will run automatically on the first time you start the server.

Otherwise, starting the server will fail if there are pending migrations, so you should run `ak server migrate` explicitly to run them.

## CI Verification
On each PR where the GORM scheme file was changed, a CI job will run to verify the migration files are synced to the GORM model. If the files are synced, the job will pass, otherwise the job will fail and the developer has to generate the relevant migrations and add them to the PR.
