#!/bin/bash

# HabitFlow Server Setup Script
# Run as root on a fresh Ubuntu 22.04 server

set -e

echo "=== HabitFlow Server Setup ==="

# Update system
echo "Updating system..."
apt update && apt upgrade -y

# Install required packages
echo "Installing packages..."
apt install -y \
    curl \
    git \
    ufw \
    fail2ban

# Install Docker
echo "Installing Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
rm get-docker.sh

# Install Docker Compose
echo "Installing Docker Compose..."
apt install -y docker-compose-plugin

# Configure firewall
echo "Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw --force enable

# Configure fail2ban
echo "Configuring fail2ban..."
systemctl enable fail2ban
systemctl start fail2ban

# Create app directory
echo "Creating app directory..."
mkdir -p /opt/habitflow
cd /opt/habitflow

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Clone your repository:"
echo "   git clone https://github.com/YOUR_USERNAME/habitflow.git ."
echo ""
echo "2. Create .env file:"
echo "   cp deploy/.env.example deploy/.env"
echo "   nano deploy/.env"
echo ""
echo "3. Get SSL certificate (replace YOUR_DOMAIN and YOUR_EMAIL):"
echo "   docker compose -f deploy/docker-compose.yml run --rm certbot certonly --webroot -w /var/www/certbot -d YOUR_DOMAIN --email YOUR_EMAIL --agree-tos"
echo ""
echo "4. Start services:"
echo "   cd deploy && docker compose up -d"
echo ""
