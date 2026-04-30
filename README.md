# IAS Automations

Backend for **IAS_HC** front-end UI. Also data ingest endpoint.

## Stack

- Go
- PosgresSQL
- IndluxDB
- Redis
- Docker

# Dev Env Setup

## Specs
- Architecture: x86-64
- OS: Ubuntu 24.04 LTS

## Start (Manual)

### 1. Connect to Linux host via VSCode
- VSCode auto-install deps on target machine

### 2. Install Go

```bash
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a ~/.profile
source /etc/profile
```
Verify:

```bash
go version
# go version go1.26.1 linux/amd64
```
### 3. Install Delve (debugger)

```bash
go install -v github.com/go-delve/delve/cmd/dlv@latest
```
### 4. Git config

```bash
git config --global user.name "YOUR_NAME"
git config --global user.email "YOUR_EMAIL"
```
### 5. VSCode Go extension

Install from marketplace

### 6. Install Redis (cache)

```bash
sudo apt-get install lsb-release curl gpg
curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
sudo chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list
sudo apt-get update
sudo apt-get install redis
```
Enable Redis:

```bash
sudo systemctl enable redis-server
sudo systemctl start redis-server
```
### 7. Install Docker

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh ./get-docker.sh
```
### 8. Post-install:

```bash
sudo groupadd docker
sudo usermod -aG docker $USER
```
Done. Go build stuff.