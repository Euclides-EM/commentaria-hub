# eScriptorium Deployment Guide

## First-time Setup

### Install Docker and Docker Compose

```bash
sudo apt update
sudo apt install -y ca-certificates curl

sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

docker compose version
```

### Allow the `euclides` user to run Docker

```bash
sudo usermod -aG docker euclides
```

### Clone eScriptorium

```bash
sudo -iu euclides
cd /srv/euclides/projects
git clone https://gitlab.com/scripta/escriptorium.git
cd escriptorium
```

### Adjust Docker Compose nginx image

Replace the upstream nginx image with a local one:

```bash
sed -i 's|registry.gitlab.com/scripta/escriptorium/nginx|escriptorium-nginx:local|g' docker-compose.yml
```

### Create the `.env` file

Create a `.env` file in the `escriptorium` directory:

```dotenv
FORCE_SCRIPT_NAME="/escriptorium"
STATIC_URL="/escriptorium/static/"
LOGIN_URL="/escriptorium/login/"
LOGOUT_URL="/escriptorium/logout/"

SESSION_COOKIE_PATH="/escriptorium/"
CSRF_COOKIE_PATH="/escriptorium/"
DOMAIN=euclides.huma-num.fr
SECRET_KEY=UD685o4lXgkv
CSRF_TRUSTED_ORIGINS=https://euclides.huma-num.fr
USE_X_FORWARDED_HOST=True
SITE_NAME=escriptorium

REDIS_HOST=redis
SQL_HOST=db
SQL_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=escriptorium

DJANGO_SU_NAME=admin
DJANGO_SU_EMAIL=admin@admin.com
DJANGO_SU_PASSWORD=O55a99uBdsvr
DJANGO_FROM_EMAIL=noreply@mydomain.com

FLOWER_BASIC_AUTH=flower:changeme
KRAKEN_TRAINING_LOAD_THREADS=8
KRAKEN_TRAINING_BATCH_SIZE=1
```

## Sanity Check with Docker Compose

### Start containers

```bash
docker compose pull
docker compose up -d
```

### Check container status

```bash
docker compose ps
```

### Check logs

```bash
docker compose logs -f
```

### Local connectivity test

```bash
curl -I http://127.0.0.1:8080/
```

### Stop containers again

```bash
docker compose down
```

## 3. Create a Systemd Service

Create the service file:

```bash
sudo vim /etc/systemd/system/escriptorium.service
```

Paste:

```ini
[Unit]
Description=eScriptorium (docker compose)
Requires=docker.service
After=docker.service network-online.target
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
User=euclides
WorkingDirectory=/srv/euclides/projects/escriptorium
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

## Configure Nginx Reverse Proxy

Create the nginx site file:

```bash
sudo vim /etc/nginx/sites-available/escriptorium
```

```nginx
server {
    listen 80;
    server_name euclides.huma-num.fr;
    
    client_max_body_size 200m;

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

# Switching from the Web App to eScriptorium

## Disable the Previous Web App

```bash
sudo systemctl list-units --type=service | grep elements
sudo systemctl stop elements-resource-box
sudo systemctl disable elements-resource-box
```

## Enable eScriptorium Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now escriptorium
sudo systemctl status escriptorium
```

## Enable the Nginx Site

```bash
sudo rm -f /etc/nginx/sites-enabled/elements-resource-box
sudo ln -s /etc/nginx/sites-available/escriptorium /etc/nginx/sites-enabled/

sudo nginx -t
sudo systemctl reload nginx
```

## Access from Browser

Open:

```
https://euclides.huma-num.fr/
```

Login using the admin credentials defined in the `.env` file:

* `DJANGO_SU_NAME`
* `DJANGO_SU_PASSWORD`
* 