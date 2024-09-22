# Include variable form the .envrc file

include .envrc

# --------------------------------------------------------------------------------------------------------------#
# HELPERS -------------------------------------------------------------------------------------------------------------#


## help:print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm 
confirm:
	@echo -n "Are you Sure? [y/N]" && read ans && [ $${ans:-N} = y ]

# --------------------------------------------------------------------------------------------------------------#
# Development
# --------------------------------------------------------------------------------------------------------------#



## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api/ -db-dsn=${CINLIM_DB_DSN}

## db/psql: connect to the postgresql database using psql
.PHONY : db/psql
db/psql:
	psql ${CINLIM_DB_DSN}

## db/migration/new: create a new set of database migration
.PHONY: db/migration/new
db/migrations/new:
	@echo 'Creating migration file for ${name}'
	migrate create  -ext sql -dir  ./migration -seq ${name} 

## db/migration/up : apply all up database migration
.PHONY: db/migration/up
db/migration/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migration  -database ${CINLIM_DB_DSN} up


#
# --------------------------------------------------------------------------------------------------------------#
# Quality CONTROL
# --------------------------------------------------------------------------------------------------------------#

.PHONY: audit

audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
.PHONY: vendor
vendor:
	@echo "Tidying and veriying module dependencies"
	go mod tidy 
	go mod verify
	@echo "Vendoring Dependencies"
	go mod vendor
	 
# --------------------------------------------------------------------------------------------------------------#
# BUILD 
# --------------------------------------------------------------------------------------------------------------# 


current_time = $(shell date +%Y-%m-%dT%H:%M:%S )
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'
.PHONY: build/api
build/api:
	@echo 'Building cmd/api'
	go build -a -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api

	 
# --------------------------------------------------------------------------------------------------------------#
# PRODUCTION 
# --------------------------------------------------------------------------------------------------------------# 

production_host_ip = '20.244.47.212'

## productin/connect : connect to the production server
.PHONY : production/connect

production/connect:
	ssh azureuser@${production_host_ip}

## production/deploy/api :deploy api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -rP  ./bin/linux_amd64/api ./migration/ azureuser@${production_host_ip}:~/application/cinlim
	ssh -t azureuser@${production_host_ip} 'migrate -path ./migration  -database $$CINLIM_DB_DSN up'

## production/configure/api.service: configure the production systemd api.service file
.PHONY: production/configure/api.service
production/configure/api.service:
	rsync -P ./remote/production/api.service azureuser@${production_host_ip}:~ 
	ssh -t azureuser@${production_host_ip} '\
	sudo mv ~/api.service /etc/systemd/system/ \
	&& sudo systemctl enable api \
	&& sudo systemctl restart api \
'

## production/configure/caddyfile : configure the production caddyfile

.PHONY: production/configure/caddyfile
production/configure/caddyfile:
	rsync -P ./remote/production/Caddyfile azureuser@${production_host_ip}:~
	ssh -t azureuser@${production_host_ip} '\
		sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy\
		'
