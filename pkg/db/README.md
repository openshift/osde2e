## DB Package Development Guide

This package interfaces with the CI/CD database that stores our job results. It is separated from the other hierarchies within the `pkg` folder because it has a different development workflow and generally is concerned with different things.

The toolchain that we use to develop this package is described in detail in [this GopherCon 2020 presentation](https://www.youtube.com/watch?v=AgHdVPSty7k). If you're going to work on this package, you should watch that presentation first.

### Schema & Migrations

The schema for our database is defined within the `./migrations/` directory. In there, you'll see a sequence of numbered migration files ending with `.{up,down}.sql`. The ones ending with `.up.sql` are applied sequentially to construct the database schema. The ones ending with `.down.sql` are applied to reverse each of the ones ending with `.up.sql`. Applying all of the up migrations results in the current database schema. Applying all of the down migrations results in a totally empty database.

We're using [`go-migrate`](https://github.com/golang-migrate/migrate/v4) and [`dockertest`](https://github.com/ory/dockertest) to manage this process. If you run into trouble, consult their documentation.

To modify the schema, use the following process:

- Add new files to the `./migrations/` directory: `<number>_<name>.up.sql` and `<number>_<name>.down.sql`. Replace `<number>` and `<name>` respectively with the next unused numeric prefix and a mnemonic name for the purpose of the migration.
- Add a new row to the test cases defined within the `migrationTests` map in `db_helpers_test.go`. This map defines a set of conditions that should be true at various times in the migration lifecycle:
    - `preup`: condition that should be true before your up migration is run.
    - `during`: condition that should be true after your up migration is run.
    - `postdown`: condition that should be true after your down migration is run.
- Use those conditional hooks to check that your migration achieves the expected data transformation and that it reverses it properly. You can insert data into the database with those hooks if needed.
- Use `go test` from within the `pkg/db` folder to check that your migrations are working. The tests use docker to spin up a real postgres container, so ensure that your local system has docker available.
- Once you're happy with your schema changes, run `go generate` in the `pkg/db` folder to ensure that all of our queries still work against your new schema. If not, fix the queries by following the instructions in [queries](#queries).

### Queries

We use [`sqlc`](https://docs.sqlc.dev/) to generate type-safe query wrapper functions for our frequently-used queries. You can find the queries themselves in the `./queries/` directory. These are processed by `sqlc` when you run `go generate` to produce the files `./models.go`, `./queries.sql.go`, and `./db.go`.

To add a new query, insert it into `./queries/queries.sql` and then run `go generate`. You may get compilation errors if your query violates the schema.

Check the go types that your query is working with. Sometimes sqlc will infer a generic and harder-to-use type like `int64` in place of a useful type like `pgtype.Interval`. You can control the behavior of how `sqlc` maps SQL types to go types by modifying `./sqlc.yaml`. See [the documentation](https://docs.sqlc.dev/en/stable/reference/config.html#per-column-type-overrides) for details. Note that nullable and non-nullable types must be mapped separately.

### Recommendations

Some thoughts that may prove useful as the database evolves:

- We haven't made any effort to perfectly normalize the database or to optimize its performance. Efforts towards that will eventually make sense as the data's scale increases, but it's too easy to spend forever trying to optimize. Optimize as needed, and no more. We can probably make huge performance gains just by adding indicies to some columns.
- `sqlc` isn't perfect. It is easily confused within nested queries, but can tolerate common table expressions pretty well. If you need a nested query and are unable to get `sqlc` to compile it, try refactoring the query to use a common table expression instead (`with subqueryresults as <subquery> select * from subqueryresults`).
