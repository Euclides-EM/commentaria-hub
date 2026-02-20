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

# Install dependencies and build the backend

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

## Install swag for API docs

```bash
sudo -iu euclides
source ~/.bashrc 
go install github.com/swaggo/swag/cmd/swag@latest
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
STORE_DIR=/srv/euclides/projects/commentaria-hub/ocrflow/store
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
git pull
go generate ./...
go build -tags nogocv -o /srv/euclides/bin/ocrflow-api ./cmd/ocrflow
exit
```

Quick test:

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

## Start the service

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now commentaria-hub-api
sudo systemctl status commentaria-hub-api
sudo journalctl -u commentaria-hub-api -n 200 --no-pager
```

Quick check that it’s running:
```bash
curl -I http://127.0.0.1:8090/ || true
curl -I http://127.0.0.1:8090/api/v1/ || true
````

## Configure Nginx Reverse Proxy

Create the nginx site file:

```bash
sudo vim /etc/nginx/sites-available/commentaria-hub-api
```

```nginx
server {
    listen 80;
    server_name euclides.huma-num.fr;

    # -----------------------------
    # commentaria-hub backend routes (strip /commentaria)
    # -----------------------------

    # Redirects for missing trailing slash
    location = /commentaria/api/v1 { return 301 /commentaria/api/v1/; }
    location = /commentaria/store/data { return 301 /commentaria/store/data/; }
    location = /commentaria/swagger { return 301 /commentaria/swagger/; }

    # /commentaria/api/v1/*  ->  http://127.0.0.1:8090/api/v1/*
    location ^~ /commentaria/api/v1/ {
        proxy_pass http://127.0.0.1:8090/api/v1/;

        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout 1800s;
        proxy_send_timeout 1800s;
    }

    # /commentaria/store/data/*  ->  http://127.0.0.1:8090/store/data/*
    location ^~ /commentaria/store/data/ {
        proxy_pass http://127.0.0.1:8090/store/data/;

        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
    
    # /commentaria/swagger/*  ->  http://127.0.0.1:8090/swagger/*
    location ^~ /commentaria/swagger/ {
        proxy_pass http://127.0.0.1:8090/swagger/;
 
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout 1800s;
        proxy_send_timeout 1800s;
    }       

    # -----------------------------
    # default: eScriptorium on /
    # -----------------------------
    location / {
        proxy_pass http://127.0.0.1:8080;

        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## Enable the Nginx Site

```bash
ls -l /etc/nginx/sites-enabled/
# run `rm -f` on any existing site that conflicts with the new one, e.g. `elements-resource-box`
sudo rm -f /etc/nginx/sites-enabled/REPLACE_WITH_EXISTING_SITE_IF_ANY
sudo ln -s /etc/nginx/sites-available/commentaria-hub-api /etc/nginx/sites-enabled/

sudo nginx -t
sudo systemctl reload nginx
```

## Access from Browser

Open:

```
http://euclides.huma-num.fr/ --> eScriptorium
http://euclides.huma-num.fr/commentaria/api/v1/health --> commentaria-hub API
```
