# CI/CD Documentation

## Overview

This project uses GitHub Actions for continuous integration and deployment to Ubuntu servers.

## Workflows

### CI Workflow (`.github/workflows/ci.yml`)

Triggered on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

Jobs:
1. **Lint** - Runs golangci-lint to check code quality
2. **Build** - Compiles the Go application and uploads the binary as an artifact
3. **Test** - Runs tests with PostgreSQL database
4. **Docker** - Builds and verifies Docker image

### CD Workflow (`.github/workflows/deploy.yml`)

Triggered on:
- Push to `main` branch
- Manual workflow dispatch

Deploys the application to Ubuntu server using Docker Compose.

## Setup Instructions

### Prerequisites

1. Ubuntu server (20.04 or later)
2. SSH access to the server
3. GitHub repository secrets configured

### 1. Server Initial Setup

Use the deployment script to set up your Ubuntu server:

```bash
DEPLOY_HOST=your-server.com ./scripts/deploy.sh setup
```

This will:
- Update system packages
- Install Docker and Docker Compose
- Configure firewall (allow ports 22, 8080)
- Install required tools (git, make, curl, wget)

### 2. Configure GitHub Secrets

Add the following secrets to your GitHub repository (Settings → Secrets and variables → Actions):

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `SERVER_HOST` | Server hostname or IP address | `memology.example.com` or `192.168.1.100` |
| `SERVER_USER` | SSH username | `ubuntu` |
| `SSH_PRIVATE_KEY` | SSH private key for authentication | Contents of `~/.ssh/id_rsa` |

#### Generating SSH Key

If you don't have an SSH key pair:

```bash
# On your local machine
ssh-keygen -t rsa -b 4096 -C "deploy@memology-backend"

# Copy public key to server
ssh-copy-id ubuntu@your-server.com

# Copy private key content for GitHub secret
cat ~/.ssh/id_rsa
```

### 3. Production Environment Variables

On the server, create or update `.env` file in the deployment directory:

```bash
ssh ubuntu@your-server.com
cd ~/memology-backend
nano .env
```

Update with production values:

```env
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<strong-password>
DB_NAME=memology
DB_SSLMODE=disable

JWT_SECRET=<generate-a-strong-secret>
JWT_ACCESS_TTL=1h
JWT_REFRESH_TTL=168h
```

**Important:** Change the database password and JWT secret to strong, unique values!

### 4. Deploy

#### Automatic Deployment

Push to the `main` branch to trigger automatic deployment:

```bash
git push origin main
```

#### Manual Deployment

Using GitHub Actions:
1. Go to Actions tab in GitHub
2. Select "CD - Deploy to Ubuntu Server" workflow
3. Click "Run workflow"

Using deployment script:
```bash
DEPLOY_HOST=your-server.com ./scripts/deploy.sh deploy
```

## Deployment Script Usage

The `scripts/deploy.sh` script provides several commands for managing deployments:

```bash
# Initial server setup
DEPLOY_HOST=memology.example.com ./scripts/deploy.sh setup

# Deploy application
DEPLOY_HOST=memology.example.com ./scripts/deploy.sh deploy

# View application logs
DEPLOY_HOST=memology.example.com ./scripts/deploy.sh logs

# Stop application
DEPLOY_HOST=memology.example.com ./scripts/deploy.sh stop

# Restart application
DEPLOY_HOST=memology.example.com ./scripts/deploy.sh restart

# Show help
./scripts/deploy.sh help
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DEPLOY_HOST` | Server hostname or IP | (required) |
| `DEPLOY_USER` | SSH username | `ubuntu` |
| `DEPLOY_DIR` | Deployment directory on server | `~/memology-backend` |
| `SSH_KEY` | Path to SSH private key | `~/.ssh/id_rsa` |

## Monitoring and Troubleshooting

### Check Application Status

```bash
ssh ubuntu@your-server.com
cd ~/memology-backend
docker-compose ps
```

### View Logs

```bash
# All logs
docker-compose logs -f

# Application logs only
docker-compose logs -f app

# PostgreSQL logs only
docker-compose logs -f postgres
```

### Restart Services

```bash
# Restart all services
docker-compose restart

# Restart app only
docker-compose restart app
```

### Health Check

Test the API endpoint:

```bash
curl http://your-server.com:8080/api/v1/auth/login
```

### Common Issues

#### 1. Port 8080 already in use

```bash
# Find process using port 8080
sudo lsof -i :8080

# Stop the process or change port in .env
```

#### 2. Database connection failed

```bash
# Check PostgreSQL container status
docker-compose logs postgres

# Verify database credentials in .env
```

#### 3. Permission denied errors

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and log back in
```

## Security Considerations

1. **SSH Keys**: Keep your SSH private keys secure. Never commit them to the repository.
2. **Secrets**: Use strong, randomly generated values for `JWT_SECRET` and database passwords.
3. **Firewall**: Only open necessary ports (22 for SSH, 8080 for API).
4. **SSL/TLS**: Consider setting up NGINX as a reverse proxy with Let's Encrypt SSL certificates.
5. **Environment Variables**: Never commit `.env` file with production credentials.

## SSL/TLS Setup (Optional but Recommended)

For production, consider adding NGINX with SSL:

```bash
# Install NGINX and Certbot
sudo apt-get install -y nginx certbot python3-certbot-nginx

# Configure NGINX as reverse proxy
sudo nano /etc/nginx/sites-available/memology

# Get SSL certificate
sudo certbot --nginx -d memology.example.com
```

Example NGINX configuration:

```nginx
server {
    listen 80;
    server_name memology.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name memology.example.com;

    ssl_certificate /etc/letsencrypt/live/memology.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/memology.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Rollback Strategy

If deployment fails:

```bash
# Stop current version
ssh ubuntu@your-server.com 'cd ~/memology-backend && docker-compose down'

# Rollback to previous commit
git checkout <previous-commit-sha>
DEPLOY_HOST=your-server.com ./scripts/deploy.sh deploy
```

## Maintenance

### Update Dependencies

```bash
# On the server
cd ~/memology-backend
docker-compose pull
docker-compose up -d
```

### Database Backup

```bash
# Backup database
docker-compose exec postgres pg_dump -U postgres memology > backup_$(date +%Y%m%d).sql

# Restore database
docker-compose exec -T postgres psql -U postgres memology < backup.sql
```

### Clean Up Old Images

```bash
# Remove unused Docker images
docker image prune -a
```

## Support

For issues or questions:
1. Check application logs: `docker-compose logs -f`
2. Review GitHub Actions workflow runs
3. Consult this documentation

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
