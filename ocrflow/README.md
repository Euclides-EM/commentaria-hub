Add a file named `.env_private` in the `ocrflow` root directory with the following content:
```
GITHUB_TOKEN=github_pat_1***5
ROBOFLOW_API_KEY=m***V
```
To get the `GITHUB_TOKEN`, follow GitHub's instructions to create a personal access token with appropriate permissions: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
To get the `ROBOFLOW_API_KEY`, sign up for a free account at Roboflow (https://roboflow.com/) and create an API key from your account settings.

Then, from the `ocrflow` root directory, run the following commands to generate necessary files and start the application:
```bash
# First time:
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/http-swagger

# Each time after that:
go generate ./...
go run ./cmd/ocrflow
```

Go to swagger UI at: http://localhost:8085/swagger/index.html
And to the viewer at: http://localhost:8085/ui/viewer/viewer.html