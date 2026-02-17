# First time setup for switching to eScriptorium

## Install Docker and Docker Compose

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

## Allow the euclides user to run Docker

```bash
sudo usermod -aG docker euclides
```

## Clone eScriptorium and create its env file

```bash
sudo -iu euclides
cd /srv/euclides/projects
git clone https://gitlab.com/scripta/escriptorium.git
cd escriptorium
```

Change the nginx image in the docker compose to `escriptorium-nginx:local`
```bash
sed -i 's|registry.gitlab.com/scripta/escriptorium/nginx|escriptorium-nginx:local|g' docker-compose.yml
```

Add an env file with the following content (adjust as needed, especially the `SECRET_KEY`):

```dotenv
FORCE_SCRIPT_NAME="/escriptorium"
STATIC_URL="/escriptorium/static/"
LOGIN_URL="/escriptorium/login/"
LOGOUT_URL="/escriptorium/logout/"

SESSION_COOKIE_PATH="/escriptorium/"
CSRF_COOKIE_PATH="/escriptorium/"
DOMAIN=euclides.huma-num.fr
SECRET_KEY=UD685o4lXgkv
CSRF_TRUSTED_ORIGINS=http://euclides.huma-num.fr
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

## Sanity - run with docker compose

```bash
docker compose pull
docker compose up -d
```
Check the status of the containers:
```bash
docker compose ps
```
Chack the logs:
```bash
docker compose logs -f
```
Sanity check locally on the server:
```bash
curl -I http://127.0.0.1:8080/
```
Take it down again:
```bash
docker compose down
```

## Create a systemd service for eScriptorium

Create:

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

## Create an nginx file for eScriptorium

```bash
sudo vim /etc/nginx/sites-available/escriptorium
```

```nginx
server {
    listen 80;
    server_name euclides.huma-num.fr;

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

# Switch to eScriptorium

## Disable the old web app

```bash
sudo systemctl list-units --type=service | grep elements
sudo systemctl stop elements-resource-box
sudo systemctl disable elements-resource-box
```

## Enable and start eScriptorium

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now escriptorium
sudo systemctl status escriptorium
```

## Enable the nginx site

```bash
sudo rm -f /etc/nginx/sites-enabled/elements-resource-box
sudo ln -s /etc/nginx/sites-available/escriptorium /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Login from browser

Navigate to http://euclides.huma-num.fr/ and login with the credentials from the env file (`DJANGO_SU_NAME`, `DJANGO_SU_PASSWORD`).