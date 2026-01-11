#!/bin/bash

# ServerEye API Production Setup Script
# Run this on your production server

set -e

echo "🚀 Setting up ServerEye API Production Server..."

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
fi

# Install Docker Compose
echo "🔧 Installing Docker Compose..."
if ! command -v docker-compose &> /dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# Create servereye directory
echo "📁 Creating application directory..."
sudo mkdir -p /opt/servereye
sudo chown $USER:$USER /opt/servereye
cd /opt/servereye

# Create .env file
echo "⚙️ Creating environment configuration..."
if [ ! -f .env ]; then
    cat > .env << EOF
# Production Configuration
HOST=0.0.0.0
PORT=8080

# Database (configure if using PostgreSQL)
DATABASE_URL=
POSTGRES_DB=servereye
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password_here

# Security (CHANGE THESE!)
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
WEBHOOK_SECRET=your-webhook-secret-min-32-chars

# Web
WEB_URL=https://your-domain.com

# Kafka (if using)
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=servereye-api
METRICS_TOPIC=metrics
EOF
    echo "✅ Created .env file - PLEASE EDIT IT with your values!"
fi

# Create logs directory
mkdir -p logs

# Download docker-compose
echo "📥 Downloading docker-compose configuration..."
curl -o docker-compose.yml https://raw.githubusercontent.com/godofphonk/ServerEyeAPI/master/docker-compose.prod.yml

# Create systemd service
echo "🔧 Creating systemd service..."
sudo cat > /etc/systemd/system/servereye.service << EOF
[Unit]
Description=ServerEye API
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/servereye
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Enable service
sudo systemctl enable servereye.service

# Setup firewall
echo "🔥 Configuring firewall..."
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw allow 8080
sudo ufw --force enable

# Setup SSL with Let's Encrypt (optional)
echo "🔒 SSL Setup (optional)..."
echo "To setup SSL with Let's Encrypt, run:"
echo "sudo apt install certbot"
echo "sudo certbot certonly --standalone -d your-domain.com"

echo ""
echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo "1. Edit /opt/servereye/.env with your configuration"
echo "2. Run: cd /opt/servereye && docker-compose up -d"
echo "3. Check: curl http://localhost:8080/health"
echo ""
echo "🌐 Your API will be available at: http://your-server-ip:8080"
echo "📊 WebSocket: ws://your-server-ip:8080/ws"
echo ""
echo "🔄 For automatic deployment, add these secrets to GitHub:"
echo "- PROD_HOST: your-server-ip"
echo "- PROD_USER: your-username"  
echo "- PROD_SSH_KEY: your-private-ssh-key"
