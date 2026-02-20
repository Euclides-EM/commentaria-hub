# SSH Deploy Key Setup for commentaria-hub

## Create a new key for commentaria-hub

Run as euclides:

```bash
sudo -iu euclides
ssh-keygen -t ed25519 -f ~/.ssh/commentaria_hub_ed25519 -C commentaria-hub@huma-num -N ""
cat ~/.ssh/commentaria_hub_ed25519.pub
```

Add that public key to GitHub:
Repo `Euclides-EM/commentaria-hub` → `Settings` → `Deploy keys` → `Add key` (read-only is fine) → paste.

## Add an SSH config so each repo uses its own key

This avoids breaking the existing deploy key.
```bash
sudo -iu euclides
cat >> ~/.ssh/config <<'EOF'

# commentaria-hub
Host github-commentaria
HostName github.com
User git
IdentityFile ~/.ssh/commentaria_hub_ed25519
IdentitiesOnly yes
EOF

chmod 600 ~/.ssh/config
```
Now clone using the alias host:

```bash
cd /srv/euclides/projects
git clone git@github-commentaria:Euclides-EM/commentaria-hub.git
```

Quick check that the alias works:

```bash
ssh -T git@github-commentaria
```

# Install dependencies and build the frontend

## Install Go

```bash
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

Check that it works:

```bash
go version
```

Add to `.bashrc`:

```bash
sudo -iu euclides
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

Check that it works:

```bash
go version
```

## Build the backend

On the server you can **build without OpenCV** (no deskew). The API will run; dataset creation will copy images without deskewing when "deskew" is requested. No need to install `libopencv-dev` or fight OpenCV/gocv version mismatches.

(Optional) If you want deskew on the server you must install **OpenCV 4.7+** (Ubuntu’s `libopencv-dev` is often older and incompatible with gocv’s ArUco bindings). Then remove the `-tags nogocv` build flag below.

As root, run:
```bash
sudo mkdir -p /srv/euclides/bin
sudo chown -R euclides:euclides /srv/euclides/bin
```

Then login as euclides and build the backend, replace the `GITHUB_TOKEN` value with the one from your `.env_private` file:
```bash
sudo -iu euclides
cd /srv/euclides/projects/commentaria-hub/ocrflow
source ~/.bashrc 
go install github.com/swaggo/swag/cmd/swag@latest
go generate ./...
go build -tags nogocv -o /srv/euclides/bin/ocrflow-api ./cmd/ocrflow
exit
```

## Add env file

```bash
sudo mkdir -p /etc/euclides
sudo chmod 700 /etc/euclides
sudo vim /etc/euclides/commentaria-hub-api.env
```

Add (minimally):
```dotenv
HTTP_ADDR=127.0.0.1:8090
STORE_DIR=/srv/euclides/data/commentaria-hub
ESCRIPTORIUM_USERNAME=admin
ESCRIPTORIUM_PASSWORD=
GITHUB_TOKEN=***
ROBOFLOW_API_KEY=***
```

Use the `GITHUB_TOKEN` and `ROBOFLOW_API_KEY` secrets from your own `.env_private` file.
Use the `ESCRIPTORIUM_USERNAME` and `ESCRIPTORIUM_PASSWORD` that you set up in the eScriptorium deployment, you can check it by running:

```bash
sudo -iu euclides
cat /srv/euclides/projects/escriptorium/variables.env | grep DJANGO_SU 
```

Create the data directory:

```bash
sudo mkdir -p /srv/euclides/data/commentaria-hub
sudo chown -R euclides:euclides /srv/euclides/data/commentaria-hub
```

## Create a systemd service

```bash
sudo vim /etc/systemd/system/commentaria-hub-api.service
```

Add the following content:'

```ini
[Unit]
Description=commentaria-hub API (ocrflow)
After=network.target

[Service]
Type=simple
User=euclides
EnvironmentFile=-/etc/euclides/commentaria-hub-api.env
ExecStart=/srv/euclides/bin/ocrflow-api
WorkingDirectory=/srv/euclides/projects/commentaria-hub/ocrflow
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
```