From the `ocrflow` root directory, run the following commands to generate necessary files and start the application:
```bash
# First time:
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/http-swagger

# Each time after that:
go generate ./...
go run ./cmd/ocrflow
```