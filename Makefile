run:
	OTEL_EXPORTER_OTLP_INSECURE="true" go run ./cmd

build:
	docker build -t lordrahl/notes:latest .

push:
	docker push lordrahl/notes

lint:
	golangci-lint run

apply-migration:
	migrate -database mysql://root:@tcp(localhost:3308)/inventory -path=$(path) up all

rollback-migration:
	migrate -database mysql://tcp(localhost:3308)/$(db) -path=$(path) down all

create-migration:
	 migrate create -ext sql -dir $(path) $(name)


am: apply-migration
rollback: rollback-migration
cm: create-migration

