# App setup

```shell
cd app
yarn
```

# Python setup

```shell
cd python-tools
uv sync
```

# Go setup

Add a file named `.env_private` in the `ocrflow` root directory with the following content:
```
GITHUB_TOKEN=github_pat_1***5
ROBOFLOW_API_KEY=m***V
```
To get the `GITHUB_TOKEN`, follow GitHub's instructions to create a personal access token with appropriate permissions: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
To get the `ROBOFLOW_API_KEY`, sign up for a free account at Roboflow (https://roboflow.com/) and create an API key from your account settings.

Then, from the `ocrflow` root directory, run the following commands to generate necessary files and start the application:

**On macOS (with deskew):** install OpenCV then build:
```bash
# First time:
brew install opencv
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/http-swagger

# Build and run (with deskew):
go generate ./...
go run ./cmd/ocrflow
```

**On Linux server (no OpenCV):** build without the `gocv` tag. No OpenCV install needed, this means you won't be able to create deskewed datasets on the server, but the API will run and you can create non-deskewed datasets.
If you want deskew on the server you must install **OpenCV 4.7+** (Ubuntu’s `libopencv-dev` is often older and incompatible with gocv’s ArUco bindings). Then remove the `-tags nogocv` build flag below.
```bash
go generate ./...
go build -tags nogocv -o /srv/euclides/bin/ocrflow-api ./cmd/ocrflow
# Optional: go build -tags gocv -o ... if you have OpenCV 4.7+ installed
```

Go to swagger UI at: http://localhost:8085/swagger/index.html
And to the viewer at: http://localhost:8085/ui/viewer/viewer.html
