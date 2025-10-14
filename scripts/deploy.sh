#!/bin/bash

# Deployment script for Memology Backend on Ubuntu Server
# This script helps with initial setup and manual deployments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DEPLOY_USER="${DEPLOY_USER:-ubuntu}"
DEPLOY_HOST="${DEPLOY_HOST:-}"
DEPLOY_DIR="${DEPLOY_DIR:-~/memology-backend}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa}"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required variables are set
check_requirements() {
    if [ -z "$DEPLOY_HOST" ]; then
        print_error "DEPLOY_HOST is not set. Usage: DEPLOY_HOST=your-server.com ./scripts/deploy.sh"
        exit 1
    fi
}

# Setup server with required dependencies
setup_server() {
    print_info "Setting up Ubuntu server with required dependencies..."
    
    ssh -i "$SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" << 'ENDSSH'
        set -e
        
        # Update system
        echo "Updating system packages..."
        sudo apt-get update
        sudo apt-get upgrade -y
        
        # Install Docker
        if ! command -v docker &> /dev/null; then
            echo "Installing Docker..."
            sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
            sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io
            
            # Add user to docker group
            sudo usermod -aG docker $USER
            echo "✅ Docker installed successfully"
        else
            echo "✅ Docker is already installed"
        fi
        
        # Install Docker Compose
        if ! command -v docker-compose &> /dev/null; then
            echo "Installing Docker Compose..."
            sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
            sudo chmod +x /usr/local/bin/docker-compose
            echo "✅ Docker Compose installed successfully"
        else
            echo "✅ Docker Compose is already installed"
        fi
        
        # Install other useful tools
        sudo apt-get install -y git make curl wget
        
        # Configure firewall
        if command -v ufw &> /dev/null; then
            echo "Configuring firewall..."
            sudo ufw allow 22/tcp
            sudo ufw allow 8080/tcp
            sudo ufw --force enable || true
            echo "✅ Firewall configured"
        fi
        
        echo "✅ Server setup completed successfully!"
        echo "⚠️  Please log out and log back in for Docker group changes to take effect"
ENDSSH

    print_info "Server setup completed!"
}

# Deploy application
deploy_app() {
    print_info "Deploying application to $DEPLOY_HOST..."
    
    # Create deployment directory
    print_info "Creating deployment directory..."
    ssh -i "$SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" "mkdir -p $DEPLOY_DIR"
    
    # Copy files to server
    print_info "Copying files to server..."
    rsync -avz --delete \
        -e "ssh -i $SSH_KEY" \
        --exclude='.git' \
        --exclude='.env' \
        --exclude='bin/' \
        --exclude='*.log' \
        --exclude='node_modules/' \
        --exclude='coverage.*' \
        ./ "$DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_DIR/"
    
    # Deploy on server
    print_info "Starting deployment on server..."
    ssh -i "$SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" << ENDSSH
        set -e
        cd $DEPLOY_DIR
        
        # Create .env file if it doesn't exist
        if [ ! -f .env ]; then
            echo "Creating .env file from example..."
            cp .env.example .env
            echo "⚠️  Please update .env with production values!"
        fi
        
        # Stop existing containers
        echo "Stopping existing containers..."
        docker-compose down || true
        
        # Pull and rebuild
        echo "Building and starting containers..."
        docker-compose pull postgres || true
        docker-compose build --no-cache app
        
        # Start services (use production override if available)
        if [ -f docker-compose.prod.yml ]; then
            docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
        else
            docker-compose up -d
        fi
        
        # Wait for services
        echo "Waiting for services to start..."
        sleep 10
        
        # Check status
        docker-compose ps
        
        echo "✅ Deployment completed!"
ENDSSH

    print_info "Checking application health..."
    sleep 5
    if curl -f "http://$DEPLOY_HOST:8080/api/v1/auth/login" &> /dev/null; then
        print_info "✅ Application is running successfully!"
    else
        print_warning "Application might still be starting up. Check logs with: ssh $DEPLOY_USER@$DEPLOY_HOST 'cd $DEPLOY_DIR && docker-compose logs -f'"
    fi
}

# Show logs
show_logs() {
    print_info "Showing application logs..."
    ssh -i "$SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" "cd $DEPLOY_DIR && docker-compose logs -f"
}

# Stop application
stop_app() {
    print_info "Stopping application..."
    ssh -i "$SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" "cd $DEPLOY_DIR && docker-compose down"
    print_info "✅ Application stopped"
}

# Restart application
restart_app() {
    print_info "Restarting application..."
    ssh -i "$SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" "cd $DEPLOY_DIR && docker-compose restart"
    print_info "✅ Application restarted"
}

# Show help
show_help() {
    cat << EOF
Memology Backend Deployment Script

Usage: DEPLOY_HOST=your-server.com ./scripts/deploy.sh [command]

Commands:
    setup       - Initial server setup (install Docker, Docker Compose, etc.)
    deploy      - Deploy application to server
    logs        - Show application logs
    stop        - Stop application
    restart     - Restart application
    help        - Show this help message

Environment Variables:
    DEPLOY_HOST     - Server hostname or IP (required)
    DEPLOY_USER     - SSH user (default: ubuntu)
    DEPLOY_DIR      - Deployment directory on server (default: ~/memology-backend)
    SSH_KEY         - Path to SSH private key (default: ~/.ssh/id_rsa)

Examples:
    # Initial server setup
    DEPLOY_HOST=memology.example.com ./scripts/deploy.sh setup
    
    # Deploy application
    DEPLOY_HOST=memology.example.com ./scripts/deploy.sh deploy
    
    # View logs
    DEPLOY_HOST=memology.example.com ./scripts/deploy.sh logs

GitHub Actions Secrets Required:
    SERVER_HOST         - Server hostname or IP
    SERVER_USER         - SSH username
    SSH_PRIVATE_KEY     - SSH private key for authentication

EOF
}

# Main
main() {
    local command="${1:-help}"
    
    case "$command" in
        setup)
            check_requirements
            setup_server
            ;;
        deploy)
            check_requirements
            deploy_app
            ;;
        logs)
            check_requirements
            show_logs
            ;;
        stop)
            check_requirements
            stop_app
            ;;
        restart)
            check_requirements
            restart_app
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
