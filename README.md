# Backend

Salesrep backend API, deployed on docker.

## Docker Compose setup

### Installation
- Install go on your computer (see version in the Dockerfile).
- Install pre-commit and golangci-lint (see links above).
- Copy the `secrets.template.json` to `secrets.json` and fill it.
  - You can find a complete version of this file in 1Password.
- Create a `bin` directory.
- Launch the DB with `docker-compose`:
```bash
docker-compose up postgres
```
- Import a dump of an existing database into your local database (preferably the staging database).
- Restart all services with `make down up logs`.
- Enjoy!

**Useful commands:**
- `make up`: starts the backend services (API, postgres, rabbit).
- `make down`: stops all services and remove the containers.
- `make logs`: see the logs of the `backend-api` service.f

### Migrations

Database migrations are handled by [golang-migrate](https://github.com/golang-migrate/migrate).

Make sure that the [JQ](https://jqlang.github.io/jq/) CLI tool is installed on your machine. It is required to parse
your `secrets.json` file and extract your database credentials.

For convenience, we provide some Makefile commands to help you manage the migrations in your local environment.
- `make migration-create name=<migration_name>`:
  - creates 2 SQL files:
    - `0000012_migration_name.up.sql` to run the migration forward.
    - `0000012_migration_name.down.sql` to rollback the migration.
    - note that the files are prefixed with a version number.
- `make migrate-up`: runs all pending migrations.
- `make migrate-up count=<number>`: run the specified number of pending migrations.
- `make migrate-down count=<number>`: rollback the specified number of migrations.

## <a id="documentation"></a>Documentation

Every route have some swagger annotations on top of the handler function.
To generate the documentation during development, you can run:
`make generate-docs`.

Then the swagger will be accessible at [http://localhost:4000/swagger/index.html](http://localhost:4000/swagger/index.html)


### Conventions

For CRUD operations, controllers, services and repositories should be named with the following verbs:
Create/Update/Upsert/Get/Delete


## Tests

Every possible response of every route should be tested. The tests are run in CI, but you can
also run them locally using `make test`. Currently, the tests aim at checking the controllers and services logics, the repositories are mocked.
Whenever you change the interface of a repository, you can re-generate the corresponding mocks by running `make generate-mocks`.


## Troubleshooting

### Sequences
If you insert data manually/from a dumb into the db, it is possible that the sequences are messed up, and you cant create
new records from gorm. To fix the records: [https://wiki.postgresql.org/wiki/Fixing_Sequences](https://wiki.postgresql.org/wiki/Fixing_Sequences)

### Migrations #1

If this is the first time you are running the migrations on an existing local database,
it is possible that the `schema_migrations` table has not been created yet.

To fix this, you will have to run the following SQL into your database:
```sql
CREATE TABLE public.schema_migrations (
    version bigint  NOT NULL,
    dirty   boolean NOT NULL
);
INSERT INTO public.schema_migrations (version, dirty) VALUES (2, false);
ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);
```
It will create the `schema_migrations` table and indicate the system that the 2 first migrations have been run. These migrations
are the creation of the whole database schema and static tables.

### Migrations #2

When a migration fails the database is declared as `dirty` in the `schema_migrations` table. Note that, as long as the DB is
dirty, the system will not allow you to run other migrations.

To fix this, you will have to:
- fix your database manually so that it comes back to a clean state
- run `make migrations-force version=<version_number>`
  - this command forces the `version` and the `dirty` state to be updated in the `schema_migrations` table.

Note that you can also do this operation manually by forcing the `version` and the `dirty` state directly
in the `schema_migrations` table in the database.

### Migrations #3

If 2 or more developers are working on migrations, they will generate conflicting version numbers on their migration files.

To prevent this, make sure to merge regularly the `develop` branch into your feature branch.

If a conflict occurs, you will have to:
- manually rename migration files so that the prefixed version number is unique.
- make sure that the order in which the migrations are run is correct.


## Known bugs

- Make sure BOT user has "maintain" rights on the repo (if not semantic release can randomly fail)
- You can't update gorm to latest version because of github.com/jackc/pgx/v5 being used in latest, while apm still uses github.com/jackc/pgx/v4
- Sometimes migration in prod seems buggy, probably a networking issue
