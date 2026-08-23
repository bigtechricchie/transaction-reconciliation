# Transaction Reconciliation

A financial operations application for reconciling internal and external transaction records, identifying discrepancies, and supporting exception investigation workflows.

## Overview

Transaction Reconciliation models a common financial-operations process: comparing transaction records maintained internally against records received from an external bank, custodian, payment processor, broker, or other financial counterparty.

The application is built around a Go REST API, persistent transaction data, deterministic reconciliation rules, and a server-rendered operations interface.

The project focuses on:

* RESTful API design using Go's `net/http`
* transaction validation and financial business rules
* exact monetary representation using integer minor units
* SQLite persistence through `database/sql`
* deterministic reconciliation and exception classification
* defensive handling of untrusted HTTP input
* parameterised SQL and explicit error handling
* dependency and software supply-chain controls
* server-rendered views using `html/template`
* operational transaction tables with filtering, sorting, searching, and pagination

## Current Status

The application is under active development.

Implemented:

* Go API service
* explicit HTTP routing with `http.ServeMux`
* API healthcheck endpoint
* HTTP method validation

Transaction persistence, reconciliation processing, and the web interface will be added incrementally as the application evolves.

## Architecture

The intended architecture is:

```text
Browser
   |
   v
Go Web Application
   |
   | HTTP / JSON
   v
Go REST API
   |
   | database/sql
   v
SQLite
```

The API owns validation, reconciliation logic, and persistence. The web application consumes the API and presents operational data through server-rendered views.

## Security Approach

The application follows a standard-library-first and security-conscious design.

Key principles include:

* external input is treated as untrusted;
* supported HTTP methods are explicitly constrained;
* SQL statements use parameterised queries;
* sensitive internal errors are not exposed directly to clients;
* third-party dependencies require justification and review;
* dependencies are vendored once introduced;
* local database files and secrets are excluded from version control;
* only synthetic financial data is used.

Security controls are introduced alongside the attack surfaces they protect rather than as unrelated boilerplate.

## Data

The application uses realistic synthetic financial transaction data designed to resemble institutional operations workflows without containing real customer, account, or banking information.

Planned reconciliation scenarios include:

* exact matches;
* amount mismatches;
* currency mismatches;
* missing internal records;
* missing external records;
* duplicate reconciliation candidates.

## Technology

Current:

* Go
* `net/http`

Planned:

* JSON
* SQLite
* `database/sql`
* Go `html/template`
* vendored third-party dependencies where required

## Local Development

Start the API:

```bash
go run ./cmd/api
```

Check the health endpoint:

```bash
curl -i http://localhost:4000/v1/healthcheck
```

Run project checks:

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```
