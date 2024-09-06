run:
	export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
	export OTEL_RESOURCE_ATTRIBUTES="service.name=notes,service.instance.id=local-run"
	export OTEL_EXPORTER_OTLP_INSECURE="true"
	go run ./cmd

build:
	docker build -t lordrahl/notes:latest .