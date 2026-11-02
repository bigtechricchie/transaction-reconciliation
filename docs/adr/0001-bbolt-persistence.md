# ADR-0001: Use bbolt for Embedded Transaction Persistence

- **Status:** Accepted
- **Date:** 2026-11-02
- **Decision owners:** Project maintainers

## Context

Transaction Reconciliation needs durable local persistence for transaction records and related reconciliation state.

The persistence layer must support:

- full CRUD operations;
- atomic updates;
- predictable local development on Windows;
- no mandatory external database server for the initial deployment model;
- minimal third-party dependency surface;
- explicit supply-chain review;
- a clean migration path if the application later requires a relational database.

Several options were evaluated before this decision.

### PostgreSQL

PostgreSQL remains the strongest long-term relational option for workloads that require:

- many concurrent writers;
- multiple application instances;
- complex joins;
- ad-hoc reporting;
- richer relational constraints;
- large-scale analytical queries.

It was not selected for the initial persistence layer because it introduces a separately operated database service and more local-development and deployment complexity than the project currently needs.

PostgreSQL remains the preferred migration target if the application's concurrency or relational requirements outgrow an embedded store.

### SQLite

SQLite provides embedded SQL and ACID transactions and would otherwise be a strong fit.

The Go driver options reviewed introduced trade-offs that did not fit this project's current dependency policy:

- some pure-Go drivers introduced a broader transitive dependency graph than desired;
- mature alternatives may require CGO and an additional native compiler toolchain on Windows.

SQLite was therefore not selected for the initial implementation.

### BadgerDB

BadgerDB provides embedded transactional key/value storage and built-in encryption options.

Its selected Go module graph was materially larger than desired for this project, so it was rejected under the same supply-chain minimisation standard applied to the SQLite candidates.

### bbolt

The reviewed version is:

~~~text
go.etcd.io/bbolt v1.5.0
~~~

A minimal consumer reduced to:

~~~text
go.etcd.io/bbolt v1.5.0
golang.org/x/sys v0.45.0
~~~

The relevant compiled packages observed on Windows were:

~~~text
go.etcd.io/bbolt/errors
go.etcd.io/bbolt/internal/common
go.etcd.io/bbolt/internal/freelist
golang.org/x/sys/windows
go.etcd.io/bbolt
~~~

This gave bbolt the smallest acceptable effective dependency surface among the embedded options reviewed.

## Decision

Use **bbolt v1.5.0** as the initial embedded persistence engine for Transaction Reconciliation.

Application code must not access bbolt directly from HTTP handlers.

Persistence must be isolated behind application-owned storage abstractions so the storage engine can be replaced later without rewriting business or HTTP logic.

The initial architecture should follow this boundary:

~~~text
HTTP handlers
    |
    v
application/service layer
    |
    v
transaction storage interface
    |
    v
bbolt implementation
    |
    v
local database file
~~~

## Why bbolt fits the current project

bbolt provides:

- pure-Go operation;
- Windows support;
- no CGO requirement;
- no separate database server;
- no Docker requirement;
- fully serializable ACID transactions;
- multiple concurrent readers;
- one write transaction at a time;
- a small effective dependency surface;
- a mature, stable embedded key/value storage model.

These properties match the current application's expected workload: structured transaction records, predictable access patterns, moderate write volume, and a single application instance.

## Concurrency model

bbolt supports many concurrent readers but only one read-write transaction at a time.

This is acceptable for the current architecture provided write transactions remain short.

Slow work must not happen inside a write transaction.

Do this:

~~~text
validate and prepare data
        |
        v
open short write transaction
        |
        v
write records and indexes
        |
        v
commit
~~~

Do not do this:

~~~text
open write transaction
        |
        v
call remote service
        |
        v
wait for network or user input
        |
        v
commit
~~~

If write contention becomes noticeable, the application may introduce a bounded application-level write queue, back-pressure, timeouts, and metrics.

Queue position should only be exposed to users if measurements show that waiting is materially noticeable.

## Data model implications

bbolt is a key/value database, not a relational database.

The application owns:

- bucket design;
- key design;
- secondary indexes;
- relationship maintenance;
- migration logic;
- consistency between primary records and indexes.

For example, the project may eventually use buckets such as:

~~~text
transactions
transactions_by_reference
transactions_by_status
transactions_by_date
reconciliation_runs
exceptions
audit_events
metadata
~~~

Any primary-record and secondary-index updates that must remain consistent should be performed inside the same bbolt write transaction.

## Financial data representation

Monetary values must continue to use integer minor units.

For example:

~~~go
AmountMinor int64
~~~

Do not introduce binary floating-point storage for transaction amounts.

## Security boundaries

bbolt provides transactional integrity and durability for its intended workload.

It does **not** provide automatic confidentiality for the database file.

If the database file is stolen, assume that unencrypted values can be recovered with sufficient tooling.

The binary file format must never be treated as encryption.

The security model therefore separates responsibilities:

~~~text
bbolt
→ transactional integrity and persistence

filesystem permissions
→ local access control

BitLocker or encrypted production volume
→ whole-disk / volume protection

application-level encryption
→ confidentiality for selected sensitive fields

password hashing
→ irreversible password storage

secret manager / protected environment
→ encryption-key ownership
~~~

## Sensitive-field encryption

Where data classification requires confidentiality beyond filesystem or volume encryption, selected values should be encrypted before being written to bbolt.

The intended primitive is an authenticated encryption construction such as **AES-256-GCM**, using Go's standard cryptography packages where appropriate.

Conceptually:

~~~text
sensitive plaintext
        |
        v
authenticated encryption
        |
        v
ciphertext
        |
        v
bbolt
~~~

The encryption key must not be stored in the same bbolt database.

Production keys should come from an external secret-management boundary or equivalent protected runtime configuration.

Only fields that require confidentiality should be encrypted at the application layer. Fields required for efficient lookup or indexing may need to remain non-sensitive, be represented by non-sensitive identifiers, or use a deliberately designed protected index.

Encryption design must be reviewed separately before real sensitive data is introduced.

## Passwords and authentication secrets

Passwords must never be stored as plaintext and should not be stored as reversible encrypted values.

If user authentication is introduced, password storage must use a suitable password-hashing function with per-password salts and appropriate work factors.

API keys, refresh tokens, recovery secrets, and similar credentials require their own storage and rotation policy.

## Database-file handling

The database file must:

- remain outside version control;
- never contain real customer or bank data in this public portfolio repository;
- use synthetic test data only;
- be protected with appropriate filesystem permissions;
- be included in backup and recovery design before production use;
- never contain encryption keys alongside encrypted values.

The current repository's ignore rules must continue to exclude local database files.

## Supply-chain policy

The accepted dependency is the specific reviewed version:

~~~text
go.etcd.io/bbolt v1.5.0
~~~

This is not permanent trust in all future bbolt versions.

Before upgrading:

1. review release notes and provenance;
2. inspect the resulting module graph;
3. inspect compiled dependencies where useful;
4. run `go mod verify`;
5. run `govulncheck ./...`;
6. run project tests and static checks;
7. review `go.mod` and `go.sum` diffs;
8. avoid automatic `@latest` upgrades in production changes.

If the dependency becomes suspect, follow the project's documented Go dependency incident-response procedure.

## Operational limits

bbolt is appropriate while the application remains within these assumptions:

- one application instance owns the database file;
- write transactions remain short;
- write contention remains modest;
- access patterns are known and indexable;
- complex ad-hoc relational reporting is not a primary requirement.

## Migration triggers

Re-evaluate PostgreSQL when one or more of these become true:

- multiple API instances need to write concurrently;
- sustained write contention becomes measurable;
- the application needs many concurrent writers;
- relational joins become common;
- reporting requirements become complex;
- users, transactions, reconciliation results, exceptions, and audit records develop rich relational relationships;
- operational requirements call for database-server replication, advanced backup tooling, or independently managed database availability.

The storage abstraction is intended to make this migration possible without coupling HTTP handlers or business logic to bbolt.

## Consequences

### Positive

- simple Windows development;
- no external database service;
- very small reviewed dependency surface;
- ACID persistence;
- straightforward deployment for a single-node service;
- clear security and migration boundaries.

### Negative

- one writer at a time;
- no SQL;
- no joins;
- indexes and relationships must be maintained by application code;
- database confidentiality is not built in;
- horizontal multi-instance writes require a different architecture.

## Review

This decision should be revisited when application scale, deployment topology, data sensitivity, or query complexity changes materially.
