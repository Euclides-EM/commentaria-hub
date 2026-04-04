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

## Install uv for Python (kraken integration)

```bash
sudo -iu euclides
source ~/.bashrc
curl -LsSf https://astral.sh/uv/install.sh | sh
uv --version
cd /srv/euclides/projects/commentaria-hub/python-tools
uv sync
```

## Install OpenCV

Note: This step is optional, you can build the backend with `-tags nogocv` to skip OpenCV and deskewing on the server. If you want deskewing, you must have a compatible OpenCV version installed (4.7+ for gocv’s ArUco bindings). Ubuntu’s `libopencv-dev` is often older, so you may need to install from source or use a PPA.

```bash
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$(go env GOPATH)/bin
cd ~
git clone https://github.com/hybridgroup/gocv.git
cd gocv
make install
```

## Install Sqlite for debugging (optional, not needed if you only use Postgres)

```bash
sudo apt-get update
sudo apt install -y sqlite3
```

You can now query using:

```bash
sudo -iu euclides
sqlite3 /srv/euclides/projects/commentaria-hub/ocrflow/store/ocrflow.db
```

## Setup the store and backup directories

Assuming the volume is mounted at `/data` and you want to use `/data/euclides/commentaria-hub/store` for the store and `/data/euclides/commentaria-hub/full_backups` for the backups.

Create those directories and set permissions as `root`:

```bash
sudo mkdir -p /data/euclides/commentaria-hub/store
sudo mkdir -p /data/euclides/commentaria-hub/full_backups
```

Set ownership to the `euclides` user so the API can read/write:

```bash
sudo chown -R euclides:euclides /data/euclides
```

Test that the `euclides` user can write to the store directory:

```bash
sudo -u euclides touch /data/euclides/commentaria-hub/store/test
sudo -u euclides ls -l /data/euclides/commentaria-hub/store
sudo -u euclides rm /data/euclides/commentaria-hub/store/test
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
ROOT_DIR=/srv/euclides/projects/commentaria-hub/ocrflow
STORE_DIR=/data/euclides/commentaria-hub/store
BACKUP_ROOT_DIR=/data/euclides/commentaria-hub/full_backups
ESCRIPTORIUM_USERNAME=admin
ESCRIPTORIUM_PASSWORD=
GITHUB_TOKEN=***
ROBOFLOW_API_KEY=***
UV_PATH=<path/to/uv/executable/if/not/in/PATH>
OPENAI_API_KEY=s***A
```

Use the `GITHUB_TOKEN` and `ROBOFLOW_API_KEY` secrets from your own `.env_private` file.
Use the `ESCRIPTORIUM_USERNAME` and `ESCRIPTORIUM_PASSWORD` that you set up in the eScriptorium deployment, you can check it by running:
Use the output of `which uv` for `UV_PATH`.
The `OPENAI_API_KEY` is only needed if you want to use the feature execution functionality with prompts.

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

Add the following content:

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

## Setup the FE permissions

```bash
sudo apt-get update
sudo apt-get install -y acl

sudo setfacl -m u:www-data:rx /srv/euclides
```

## Configure Nginx Reverse Proxy

Create the nginx site file:

```bash
sudo vim /etc/nginx/sites-available/commentaria-hub-api
```

```nginx
server {
    listen 80;
    server_name euclides.huma-num.fr;
    
    client_max_body_size 200m;

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
    # commentaria-hub & resource box FE apps
    # -----------------------------
    location = /hub { return 301 /hub/; }
    location ^~ /hub/ {
        alias /srv/euclides/projects/commentaria-hub/app/commentaria-app/dist/;
        try_files $uri $uri/ /hub/index.html;
    }
    
    location = /resourcebox { return 301 /resourcebox/; }
    location ^~ /resourcebox/ {
        alias /srv/euclides/projects/commentaria-hub/app/elements-resource-box/dist/;
        try_files $uri $uri/ /resourcebox/index.html;
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

## Set up SSL with Let’s Encrypt (Certbot)

```bash
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx
sudo certbot --nginx -d euclides.huma-num.fr
``` 

Note: This will automatically obtain and install the SSL certificate, and set up automatic renewal. You can test the renewal process with:

```bash
sudo certbot renew --dry-run
```

In addition, your nginx configuration will be updated to redirect HTTP to HTTPS, and the `server_name` directive will be updated to include the SSL configuration.

Likely, the following the `listen 80;` will be removed from the existing server block and replaced by the following:

```nginx
    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/euclides.huma-num.fr/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/euclides.huma-num.fr/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
```

In addition, a new server block will be added to redirect HTTP to HTTPS:

```nginx
server {
    if ($host = euclides.huma-num.fr) {
        return 301 https://$host$request_uri;
    } # managed by Certbot


    listen 80;
    server_name euclides.huma-num.fr;
    return 404; # managed by Certbot
}
```

# Redeploying

```bash
sudo -iu euclides
cd /srv/euclides/projects/commentaria-hub/ocrflow
source ~/.bashrc
git pull
go generate ./...
go build -o /srv/euclides/bin/ocrflow-api ./cmd/ocrflow # add -tags nogocv if you built without OpenCV
cd /srv/euclides/projects/commentaria-hub/app
yarn
yarn build:euclides:huma-num
exit

sudo systemctl restart commentaria-hub-api
sudo systemctl status commentaria-hub-api
sudo journalctl -u commentaria-hub-api -n 200 --no-pager -f
```

If you update the Python dependencies (under `python-tools`), you must also run `uv sync` before building the Go backend, since the Go code calls the Python code for dataset creation. You can run `uv sync` as euclides, it will use the existing virtual environment created during setup.