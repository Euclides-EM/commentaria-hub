To set git on the Huma-Num server I followed these steps, as specified in [Github deploy leys](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/managing-deploy-keys#deploy-keys):
After SSHing into the server, I ran:
```bash
ssh-keygen -t ed25519 -C main-server@euclides.huma-num.fr # No passphrase!
cat ~/.ssh/id_ed25519.pub
``` 
Then I copied the output and added it as a deploy key in the Github repository settings, with write access.
This allowed me to clone the repository on the server.

I cloned the repository:
```bash
mkdir projects
cd projects
git clone https://github.com/Euclides-EM/elements-resource-box.git
```

Then I installed node and yarn:
```bash
curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash
source ~/.bashrc
nvm install --lts
node -v # output: v24.11.1
npm -v # output: 11.6.2
corepack enable
corepack prepare yarn@stable --activate
yarn --version # output: 4.11.0
```

I installed the dependencies:
```bash
cd elements-resource-box
yarn install
```

Then, before starting the server for the first time, I did a small hack:
```bash
vim vite-plugins/facsimile-listing.js
```
I commented out the line specifying `GITHUB_PAT` and added the lone from my local machine, which has hard coded my token.

To make it work, I did some port redirection, note that I didn't add it permanently, so after a reboot I will have to do it again:
```aiignore
sudo iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 5173
sudo iptables -t nat -A OUTPUT      -p tcp --dport 80 -j REDIRECT --to-port 5173  # for local curls
```

I set the git username and email. It's not strictly necessary, but it's better to have it set if I want to do commits from the server:
```bash
git config user.email "main-server@euclides.huma-num.fr"
git config user.name 'Euclides HumaNum Server'
```

And finally, I started the development server:
```bash
yarn dev --host --port 5173
```

After I checked it's OK, I run it in nohup:
For the first time, I had to create the log folder:
```bash
sudo mkdir -p /var/log/elements-resource-box
sudo chown $USER:$USER /var/log/elements-resource-box/
```
And then I ran:
```bash
# to self, try to use this next time: $(date +'%Y-%m-%d_%H-%M')_out.log
nohup yarn dev --host --port 5173 > /var/log/elements-resource-box/out.log 2>/var/log/elements-resource-box/err.log &
```
To check the logs:
```bash
tail -f /var/log/elements-resource-box/out.log
tail -f /var/log/elements-resource-box/err.log
```
To check the running processes:
```bash
ps aux | grep yarn
```
To stop the server, I ran:
```bash
kill <PID> # or maybe pkill -f "yarn dev"
```

