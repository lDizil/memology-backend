# Quick Start - Deployment

## One-Time Setup

### 1. Configure GitHub Secrets

Navigate to: **Settings → Secrets and variables → Actions → New repository secret**

Add these secrets:
- `SERVER_HOST` - Your server IP or domain (e.g., `192.168.1.100` or `memology.example.com`)
- `SERVER_USER` - SSH username (usually `ubuntu`)
- `SSH_PRIVATE_KEY` - Your SSH private key content

### 2. Setup Server

```bash
# Generate SSH key if you don't have one
ssh-keygen -t rsa -b 4096 -C "deploy@memology"

# Copy public key to server
ssh-copy-id ubuntu@YOUR_SERVER_IP

# Run server setup
DEPLOY_HOST=YOUR_SERVER_IP ./scripts/deploy.sh setup
```

### 3. Configure Production Environment

```bash
# SSH to server
ssh ubuntu@YOUR_SERVER_IP

# Navigate to deployment directory
cd ~/memology-backend

# Create .env file
cp .env.example .env
nano .env

# Update these values:
# - DB_PASSWORD (use a strong password)
# - JWT_SECRET (generate with: openssl rand -base64 32)
```

## Deployment Methods

### Automatic (Recommended)
Push to `main` branch:
```bash
git push origin main
```

### Manual via GitHub
1. Go to **Actions** tab
2. Select **CD - Deploy to Ubuntu Server**
3. Click **Run workflow**

### Manual via Script
```bash
DEPLOY_HOST=YOUR_SERVER_IP ./scripts/deploy.sh deploy
```

## Common Commands

```bash
# View logs
DEPLOY_HOST=YOUR_SERVER_IP ./scripts/deploy.sh logs

# Restart application
DEPLOY_HOST=YOUR_SERVER_IP ./scripts/deploy.sh restart

# Stop application
DEPLOY_HOST=YOUR_SERVER_IP ./scripts/deploy.sh stop
```

## Health Check

```bash
# Test API
curl http://YOUR_SERVER_IP:8080/api/v1/auth/login

# Expected response: Method Not Allowed (this is OK - endpoint exists)
```

## Troubleshooting

See [docs/CICD.md](CICD.md) for detailed troubleshooting guide.
