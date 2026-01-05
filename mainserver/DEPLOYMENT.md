# Deployment Guide (Huma-Num server)

## Goal

* Multiple web apps on a server where only **port 80** is open
* No manual redeploy after reboot
* Predictable start, stop, restart during deploys

Approach:

* **Nginx** listens on port 80 and routes traffic by hostname
* Each app runs on a **local port** (127.0.0.1:5173, 127.0.0.1:5174, etc)
* Each app is managed by **systemd** (auto-start on boot, auto-restart, logs via journalctl)
* Secrets are injected via **env files**, not hardcoded in repo files
* No **iptables redirect** rules are used

---

## Conventions used in this guide

* Projects live under: `/srv/euclides/projects`
* Service user: `euclides`
* App 1 example:

    * repo: `elements-resource-box`
    * local port: `5173`
    * domain: `elements.example.org`
    * systemd service: `elements-resource-box`
* App 2 example:

    * repo: `other-app`
    * local port: `5174`
    * domain: `other.example.org`
    * systemd service: `other-app`

Adjust names, ports, domains as needed.

---

# First-time setup (server-wide)

## 1) Install base packages

```bash
sudo apt update
sudo apt install -y nginx git curl
sudo systemctl enable --now nginx
```

## 2) Create a dedicated service user

```bash
sudo adduser --system --group --home /srv/euclides euclides
sudo mkdir -p /srv/euclides/projects
sudo chown -R euclides:euclides /srv/euclides
sudo chsh -s /bin/bash euclides
```

## 3) Install Node via nvm for the service user

Switch to the service user and install Node + Yarn once:

```bash
sudo -iu euclides
touch ~/.bashrc
touch ~/.profile
curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash
source ~/.bashrc
nvm install --lts
node -v
npm -v
corepack enable
corepack prepare yarn@stable --activate
yarn --version
exit
```

Notes:

* Node and Yarn live under the `euclides` user’s home
* Avoid relying on root’s environment for runtime services

---

# First-time setup (per repository)

## 4) Configure GitHub deploy key (recommended)

Run as the service user:

```bash
sudo -iu euclides
ssh-keygen -t ed25519 -C main-server@euclides.huma-num.fr
cat ~/.ssh/id_ed25519.pub
```

Add the public key as a **Deploy key** in GitHub repo settings.

* Prefer **read-only** deploy keys unless server-side commits are truly needed

Optional git identity (only needed for commits from server):

```bash
git config --global user.email "main-server@euclides.huma-num.fr"
git config --global user.name "Euclides HumaNum Server"
```

## 5) Clone the repo

```bash
sudo -iu euclides
cd /srv/euclides/projects
git clone git@github.com:Euclides-EM/elements-resource-box.git
cd elements-resource-box
yarn install
exit
```

---

# First-time setup (secrets, per app)

## 6) Put secrets in an env file (no hardcoding in JS files)

Create a folder for env files:

```bash
sudo mkdir -p /etc/euclides
sudo chown root:root /etc/euclides
sudo chmod 700 /etc/euclides
```

Create an env file for this app:

```bash
sudo vim /etc/euclides/elements-resource-box.env
```

Example contents:

```bash
GITHUB_PAT=xxxxx
```

[//]: # (Lock it down:)

[//]: # ()
[//]: # (```bash)

[//]: # (sudo chown root:root /etc/euclides/elements-resource-box.env)

[//]: # (sudo chmod 600 /etc/euclides/elements-resource-box.env)

[//]: # (```)

[//]: # ()
[//]: # (Code should read token from environment, eg `process.env.GITHUB_PAT`.)

[//]: # (Once this exists, remove the “vim hack” that hardcodes tokens into repo files.)

---

# First-time setup (systemd, per app)

## 7) Create a systemd service for the app

Create:

```bash
sudo vim /etc/systemd/system/elements-resource-box.service
```

Paste:

```ini
[Unit]
Description=elements-resource-box (Vite)
After=network.target

[Service]
Type=simple
User=euclides
WorkingDirectory=/srv/euclides/projects/elements-resource-box
Environment=HOST=127.0.0.1
Environment=PORT=5173
EnvironmentFile=-/etc/euclides/elements-resource-box.env
ExecStart=/bin/bash -c "source ~/.bashrc && yarn dev --host ${HOST} --port ${PORT}"
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now elements-resource-box
sudo systemctl status elements-resource-box
```

Logs:

```bash
sudo journalctl -u elements-resource-box -f
```

---

# First-time setup (Nginx routing, per app)

## 8) Add an Nginx site for the app (hostname routing)

Create:

```bash
sudo vim /etc/nginx/sites-available/elements-resource-box
```

Paste:

```nginx
server {
    listen 80;
    server_name elements.example.org;

    location / {
        proxy_pass http://127.0.0.1:5173;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/elements-resource-box /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

# Regular operations (day-to-day)

These are the commands used for stop, start, restart, deploy.

## Check status

```bash
sudo systemctl status elements-resource-box
```

## View logs

Follow logs:

```bash
sudo journalctl -u elements-resource-box -f
```

Last 200 lines:

```bash
sudo journalctl -u elements-resource-box -n 200 --no-pager
```

## Stop

```bash
sudo systemctl stop elements-resource-box
```

## Start

```bash
sudo systemctl start elements-resource-box
```

## Restart (most common after deploy)

```bash
sudo systemctl restart elements-resource-box
```

---

# Deploying a new version (typical flow)

## Redeploy steps (repeat every deploy)

```bash
sudo -iu euclides
cd /srv/euclides/projects/elements-resource-box
git pull
yarn install
exit

sudo systemctl restart elements-resource-box
sudo journalctl -u elements-resource-box -n 80 --no-pager
```

If a secret changed:

```bash
sudo nano /etc/euclides/elements-resource-box.env
sudo systemctl restart elements-resource-box
```

---

# Adding a second web app

Repeat the per-repo steps, with these differences:

* New folder: `/srv/euclides/projects/other-app`
* New local port: `5174` (or any free one)
* New env file: `/etc/euclides/other-app.env`
* New service: `/etc/systemd/system/other-app.service`
* New Nginx site: `/etc/nginx/sites-available/other-app`
* New hostname: `other.example.org`

Nginx proxy_pass should target the second port:

```nginx
proxy_pass http://127.0.0.1:5174;
```
