run:
	OTEL_EXPORTER_OTLP_INSECURE="true" go run ./cmd

build:
	docker build -t lordrahl/notes:latest .

push:
	docker push lordrahl/notes