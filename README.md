
# Cinlim Movie Management Server

This project is a Go-based server for managing a movie database, utilizing PostgreSQL as the database engine. The server includes features such as rate limiting, security headers, email verification, performance metrics, and a graceful shutdown process for enhanced stability and security. It also includes deployment scripts for production using Caddy and Azure.

## Prerequisites

Before running the commands in this Makefile, ensure that you have the following:
- Go installed
- PostgreSQL installed and running
- `migrate` tool installed for database migrations
- SSH access to the production server
- `.envrc` file with necessary environment variables

---

## Makefile Commands Overview

### Development

#### Run API
```bash
make run/api
```
Runs the `cmd/api` application using the database credentials specified in the `.envrc` file.

#### Connect to PostgreSQL Database
```bash
make db/psql
```
Connects to the PostgreSQL database using `psql` with the DSN defined in the `.envrc` file.

#### Create New Database Migration
```bash
make db/migration/new name=<migration_name>
```
Creates a new SQL migration file in the `migration` directory.

#### Apply Database Migrations
```bash
make db/migration/up
```
Applies all up migrations to the PostgreSQL database. Confirmation is required.

---

### Quality Control

#### Code Audit
```bash
make audit
```
Formats the code, vets it, and runs tests for the project, including running `staticcheck`.

#### Manage Dependencies
```bash
make vendor
```
Tidies and verifies Go module dependencies, then vendors them for the project.

---

### Build

#### Build API
```bash
make build/api
```
Builds the API for both the local environment and Linux (amd64), including versioning information from Git.

---

### Production

#### Connect to Production Server
```bash
make production/connect
```
SSH into the production server using the specified production IP.

#### Deploy API to Production
```bash
make production/deploy/api
```
Deploys the API and runs database migrations on the production server.

#### Configure `api.service` on Production
```bash
make production/configure/api.service
```
Uploads and configures the `api.service` file for systemd on the production server.

#### Configure Caddy for Production
```bash
make production/configure/caddyfile
```
Uploads and configures the Caddyfile for the reverse proxy on the production server.

---

## Notes

- Ensure the `.envrc` file contains the necessary environment variables like `CINLIM_DB_DSN`.
- The production IP (`20.244.47.212`) is a placeholder and should be updated with the actual IP for your deployment environment.

