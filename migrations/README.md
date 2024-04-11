# Migrations

We use atalsgo (http://atlasgo.io/) to generate migrations based on gorm's models

## How to generate a new migration
- Update the db models that you need
- Run make generate-migrations, give a meaningful name to the migration (i.e. add-created-at-to-projects)

### Resolving conflicts in migrations
In case there are migrations created on two different branches there would be a conflict which should be resolved

First branch merging to main would work fine. Second branch would have to rebase from main and run ```atlas migrate rebase <versions>``` from the problematic versions
. Consule [Atlas Docs](https://atlasgo.io/versioned/apply) for further explanations.

## Running migrations
if ```AK_DB__AUTO_MIGRATE=true``` migrations would run automatically first time you start the server


otherwise, starting the server would fail if there are pending migrations and you should run ```ak server migrate``` explicitly to run them


## CI Verification
On each PR where gorm scheme file was changed, a CI job would run to verify the migration files are synced to the gorm model, if the files are synced, job would pass, otherwise, the job fail and the developer has to generate the relevant migrations and add them to the PR
