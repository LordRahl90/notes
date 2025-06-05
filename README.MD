# Notes

This is a simple go application to test k-native deployment. <br />

It creates a note and stores it into a map. So you can lose the data once a new version is deployed <br />

This is meant to test `otel-lgtm` and some other external services.

# Start"

RUN `cp .env.example .env` to set the environment variables.
# Run
To run the application, you can use the following command:

```bash 
go run main.go
``` 
OR 
```bash
go run ./cmd/
``` 
OR
```bash
make run
```

## Note
The `docker-compose.yml` file is used to run the application in a containerized environment.
You can use the following command to start the application using Docker Compose:
```bash
docker compose up
```
