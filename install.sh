#!/bin/bash

# 0G Galileo Prometheus Monitoring System Installation Script
# This script installs all required dependencies and sets up the monitoring system

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if port is in use
port_in_use() {
    # Check if port is in use by any process
    if lsof -i :$1 >/dev/null 2>&1; then
        return 0
    fi
    
    # Check if port is in use by Docker containers
    if docker ps --format "table {{.Ports}}" | grep -q ":$1->" 2>/dev/null; then
        return 0
    fi
    
    return 1
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    print_status "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            print_success "$service_name is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within expected time"
    return 1
}

# Main installation function
main() {
    echo "=========================================="
    echo "  0G Galileo Prometheus Monitoring System"
    echo "           Installation Script"
    echo "=========================================="
    echo ""
    
    # Check if running as root
    if [ "$EUID" -eq 0 ]; then
        print_warning "Running as root. This is not recommended for security reasons."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_error "Installation aborted."
            exit 1
        fi
    fi
    
    # Update system packages
    print_status "Updating system packages..."
    sudo apt update
    
    # Install essential packages
    print_status "Installing essential packages..."
    sudo apt install -y curl wget git unzip software-properties-common apt-transport-https ca-certificates gnupg lsb-release
    
    # Install Docker
    print_status "Installing Docker..."
    if ! command_exists docker; then
        # Remove old versions
        sudo apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true
        
        # Add Docker's official GPG key
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        
        # Add Docker repository
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        
        # Install Docker
        sudo apt update
        sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    else
        print_success "Docker is already installed: $(docker --version)"
        
        # Check if Docker service is running
        if ! sudo systemctl is-active --quiet docker; then
            print_status "Starting Docker service..."
            sudo systemctl start docker
            sudo systemctl enable docker
        fi
    fi
    
    # Install Docker Compose (standalone)
    print_status "Installing Docker Compose..."
    if ! command_exists docker-compose; then
        # Download Docker Compose
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        
        # Create symlink
        sudo ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    else
        print_success "Docker Compose is already installed: $(docker-compose --version)"
    fi
    
    # Add user to docker group
    if [ "$EUID" -ne 0 ]; then
        print_status "Adding user to docker group..."
        sudo usermod -aG docker $USER
        print_warning "You need to log out and log back in for docker group changes to take effect."
    fi
    
    # Install Node Exporter
    print_status "Installing Node Exporter..."
    if ! command_exists node_exporter; then
        # Download Node Exporter
        NODE_EXPORTER_VERSION="1.6.1"
        NODE_EXPORTER_URL="https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz"
        
        cd /tmp
        wget $NODE_EXPORTER_URL
        tar xvfz node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz
        sudo cp node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64/node_exporter /usr/local/bin/
        sudo chmod +x /usr/local/bin/node_exporter
        
        # Create systemd service for Node Exporter
        sudo tee /etc/systemd/system/node_exporter.service > /dev/null <<EOF
[Unit]
Description=Node Exporter
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/node_exporter --web.listen-address=:9200
Restart=always

[Install]
WantedBy=multi-user.target
EOF
        
        # Enable and start Node Exporter
        sudo systemctl daemon-reload
        sudo systemctl enable node_exporter
        sudo systemctl start node_exporter
        
        # Cleanup
        rm -rf node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64*
    else
        print_success "Node Exporter is already installed"
        
        # Check if Node Exporter is running as a process
        if pgrep -x "node_exporter" > /dev/null; then
            print_status "Node Exporter is running as a process"
            
            # Check if systemd service exists
            if [ ! -f "/etc/systemd/system/node_exporter.service" ]; then
                print_status "Creating systemd service for existing Node Exporter..."
                
                # Find Node Exporter executable path
                NODE_EXPORTER_PATH=$(which node_exporter 2>/dev/null || echo "/usr/local/bin/node_exporter")
                
                # If not found in PATH, try to find it in common locations
                if [ ! -f "$NODE_EXPORTER_PATH" ]; then
                    if [ -f "/root/node_exporter-1.5.0.linux-amd64/node_exporter" ]; then
                        NODE_EXPORTER_PATH="/root/node_exporter-1.5.0.linux-amd64/node_exporter"
                    elif [ -f "/usr/bin/node_exporter" ]; then
                        NODE_EXPORTER_PATH="/usr/bin/node_exporter"
                    else
                        print_warning "Could not find Node Exporter executable. Please check installation."
                        NODE_EXPORTER_PATH="/usr/local/bin/node_exporter"
                    fi
                fi
                
                # Create systemd service
                sudo tee /etc/systemd/system/node_exporter.service > /dev/null <<EOF
[Unit]
Description=Node Exporter
After=network.target

[Service]
Type=simple
User=root
ExecStart=$NODE_EXPORTER_PATH --web.listen-address=:9200
Restart=always

[Install]
WantedBy=multi-user.target
EOF
                
                # Stop existing process and start as service
                print_status "Stopping existing Node Exporter process..."
                pkill -f "node_exporter" || true
                sleep 2
                
                # Enable and start service
                sudo systemctl daemon-reload
                sudo systemctl enable node_exporter
                sudo systemctl start node_exporter
                
                print_success "Node Exporter converted to systemd service"
            else
                print_status "Node Exporter systemd service already exists"
            fi
        fi
        
        # Check if Node Exporter service is running
        if ! sudo systemctl is-active --quiet node_exporter; then
            print_status "Starting Node Exporter service..."
            sudo systemctl start node_exporter
            sudo systemctl enable node_exporter
        fi
    fi
    
    # Install Go (if not already installed)
    print_status "Installing Go..."
    if ! command_exists go; then
        GO_VERSION="1.21.13"
        GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
        
        cd /tmp
        wget $GO_URL
        sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
        
        # Add Go to PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
        export PATH=$PATH:/usr/local/go/bin
        
        # Cleanup
        rm go${GO_VERSION}.linux-amd64.tar.gz
    else
        print_success "Go is already installed: $(go version)"
    fi
    
    # Configure firewall
    print_status "Configuring firewall..."
    if command_exists ufw; then
        sudo ufw allow 22/tcp  # SSH
        sudo ufw allow 80/tcp  # HTTP
        sudo ufw allow 3000/tcp  # Grafana
        sudo ufw allow 8080/tcp  # Unified Metrics
        sudo ufw allow 9090/tcp  # Prometheus
        sudo ufw allow 9200/tcp  # Node Exporter
        
        # Enable firewall if not already enabled
        if ! sudo ufw status | grep -q "Status: active"; then
            print_warning "Enabling UFW firewall..."
            echo "y" | sudo ufw enable
        fi
    else
        print_warning "UFW not found. Please configure firewall manually."
    fi
    
    # Check if we're in the project directory
    if [ ! -f "docker-compose.yml" ] || [ ! -f "start.sh" ]; then
        print_error "This script must be run from the 0g_prometheus project directory."
        print_error "Please navigate to the project directory and run this script again."
        exit 1
    fi
    
    # Make scripts executable
    print_status "Making scripts executable..."
    chmod +x start.sh
    
    # Check for port conflicts
    print_status "Checking for port conflicts..."
    local ports=(80 3000 8080 9090 9200)
    local conflicts=()
    local conflict_details=()
    
    for port in "${ports[@]}"; do
        if port_in_use $port; then
            conflicts+=($port)
            # Get process details
            local process_info=$(lsof -i :$port 2>/dev/null | head -2 | tail -1 | awk '{print $1, $2}' 2>/dev/null || echo "Unknown process")
            
            # Get Docker container details
            local docker_info=$(docker ps --format "table {{.Names}}:{{.Ports}}" | grep ":$port->" 2>/dev/null | head -1 | awk '{print $1}' 2>/dev/null || echo "")
            
            if [ ! -z "$docker_info" ]; then
                conflict_details+=("Port $port: Docker container ($docker_info)")
            else
                conflict_details+=("Port $port: $process_info")
            fi
        fi
    done
    
    if [ ${#conflicts[@]} -ne 0 ]; then
        print_warning "The following ports are already in use: ${conflicts[*]}"
        echo ""
        print_status "Port conflict details:"
        for detail in "${conflict_details[@]}"; do
            echo "   $detail"
        done
        echo ""
        print_warning "This may cause conflicts with the monitoring system."
        echo "Options:"
        echo "   1. Continue anyway (may fail if ports are critical)"
        echo "   2. Stop conflicting processes and continue"
        echo "   3. Stop all Docker containers and continue"
        echo "   4. Abort installation"
        echo ""
        read -p "Choose option (1/2/3/4): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[2]$ ]]; then
            print_status "Stopping conflicting processes..."
            for port in "${conflicts[@]}"; do
                local pids=$(lsof -ti:$port 2>/dev/null)
                if [ ! -z "$pids" ]; then
                    echo "   Stopping processes on port $port: $pids"
                    kill -9 $pids 2>/dev/null || true
                fi
            done
            sleep 3
        elif [[ $REPLY =~ ^[3]$ ]]; then
            print_status "Stopping all Docker containers..."
            docker stop $(docker ps -q) 2>/dev/null || true
            docker rm $(docker ps -aq) 2>/dev/null || true
            print_status "Removing Docker networks..."
            docker network prune -f 2>/dev/null || true
            sleep 3
        elif [[ ! $REPLY =~ ^[1]$ ]]; then
            print_error "Installation aborted."
            exit 1
        fi
    fi
    
    # Test Docker installation
    print_status "Testing Docker installation..."
    if sudo docker run --rm hello-world >/dev/null 2>&1; then
        print_success "Docker is working correctly"
    else
        print_error "Docker test failed. Please check Docker installation."
        exit 1
    fi
    
    # Test Docker Compose installation
    print_status "Testing Docker Compose installation..."
    if docker-compose --version >/dev/null 2>&1; then
        print_success "Docker Compose is working correctly"
    else
        print_error "Docker Compose test failed. Please check Docker Compose installation."
        exit 1
    fi
    
    # Wait for Node Exporter to be ready
    wait_for_service "http://localhost:9200/metrics" "Node Exporter"
    
    # Build and start services
    print_status "Building and starting monitoring services..."
    if ./start.sh; then
        print_success "Monitoring services started successfully!"
    else
        print_error "Failed to start monitoring services."
        exit 1
    fi
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    
    # Wait for Unified Metrics
    if ! wait_for_service "http://localhost:8080/health" "Unified Metrics"; then
        print_warning "Unified Metrics may not be fully ready. Check logs with: docker-compose logs unified-metrics"
    fi
    
    # Wait for Grafana
    if ! wait_for_service "http://localhost:3000" "Grafana"; then
        print_warning "Grafana may not be fully ready. Check logs with: docker-compose logs grafana"
    fi
    
    # Wait for Prometheus
    if ! wait_for_service "http://localhost:9090" "Prometheus"; then
        print_warning "Prometheus may not be fully ready. Check logs with: docker-compose logs prometheus"
    fi
    
    # Wait for nginx
    if ! wait_for_service "http://localhost" "nginx"; then
        print_warning "nginx may not be fully ready. Check logs with: docker-compose logs nginx"
    fi
    
    # Display final information
    echo ""
    echo "=========================================="
    print_success "Installation completed successfully!"
    echo "=========================================="
    echo ""
    echo "📊 Access Information:"
    echo "====================="
    echo ""
    echo "🔗 Direct Access (by port):"
    echo "   Unified Metrics: http://localhost:8080/all-metrics"
    echo "   Grafana:         http://localhost:3000 (admin/admin123)"
    echo "   Prometheus:      http://localhost:9090"
    echo "   Node Exporter:   http://localhost:9200"
    echo ""
    echo "🌐 nginx Reverse Proxy:"
    echo "   Unified Metrics: http://localhost/all-metrics/"
    echo "   Grafana:         http://localhost/grafana/"
    echo "   Prometheus:      http://localhost/prometheus/"
    echo "   Node Exporter:   http://localhost/node-exporter/"
    echo ""
    echo "🌍 External Access (use your server IP):"
    echo "   Unified Metrics: http://YOUR_SERVER_IP/all-metrics/"
    echo "   Grafana:         http://YOUR_SERVER_IP/grafana/"
    echo "   Prometheus:      http://YOUR_SERVER_IP/prometheus/"
    echo ""
    echo "🔍 Quick Tests:"
    echo "=============="
    echo "   # Check unified metrics"
    echo "   curl http://localhost:8080/all-metrics | grep og_galileo"
    echo ""
    echo "   # Check beacon chain metrics"
    echo "   curl http://localhost:8080/all-metrics | grep beacon_block_signed"
    echo ""
    echo "   # Check system metrics"
    echo "   curl http://localhost:8080/all-metrics | grep node_cpu"
    echo ""
    echo "📋 Management Commands:"
    echo "======================"
    echo "   # Start services"
    echo "   ./start.sh"
    echo ""
    echo "   # Stop services"
    echo "   docker-compose down"
    echo ""
    echo "   # View logs"
    echo "   docker-compose logs -f unified-metrics"
    echo ""
    echo "   # Restart services"
    echo "   docker-compose restart"
    echo ""
    echo "   # Update services"
    echo "   docker-compose pull && docker-compose up -d"
    echo ""
    echo "⚠️  Important Notes:"
    echo "=================="
    echo "   - Beacon chain signing status is based on -1 block LastCommit info"
    echo "   - Current block N signing status = Block N-1 signing information"
    echo "   - Grafana default credentials: admin/admin123"
    echo "   - Change default passwords in production environment"
    echo "   - Node Exporter runs on port 9200 as systemd service"
    echo "   - All services are configured for auto-restart"
    echo ""
    echo "🔧 Next Steps:"
    echo "============="
    echo "   1. Access Grafana at http://localhost:3000"
    echo "   2. Import the provided dashboards"
    echo "   3. Configure alerts if needed"
    echo "   4. Set up SSL certificates for production use"
    echo "   5. Change default passwords"
    echo ""
    echo "🛠️  Troubleshooting:"
    echo "=================="
    echo "   # Check service status"
    echo "   docker-compose ps"
    echo ""
    echo "   # Check Node Exporter status"
    echo "   sudo systemctl status node_exporter"
    echo ""
    echo "   # Check Docker status"
    echo "   sudo systemctl status docker"
    echo ""
    echo "   # View all logs"
    echo "   docker-compose logs"
    echo ""
    echo "   # Check port usage"
    echo "   lsof -i :8080 -i :3000 -i :9090 -i :9200"
    echo ""
    
    # Check if user needs to log out for docker group
    if [ "$EUID" -ne 0 ] && ! groups $USER | grep -q docker; then
        print_warning "You need to log out and log back in for docker group changes to take effect."
        print_warning "Alternatively, you can run: newgrp docker"
    fi
    
    # Final status check
    echo "📊 Final Status Check:"
    echo "====================="
    if command_exists docker-compose; then
        echo "   Docker Compose: ✅ Available"
        if [ -f "docker-compose.yml" ]; then
            echo "   Project Config: ✅ Found"
        else
            echo "   Project Config: ❌ Missing"
        fi
    else
        echo "   Docker Compose: ❌ Not available"
    fi
    
    if sudo systemctl is-active --quiet node_exporter; then
        echo "   Node Exporter: ✅ Running (systemd)"
    elif pgrep -x "node_exporter" > /dev/null; then
        echo "   Node Exporter: ⚠️  Running (process)"
    else
        echo "   Node Exporter: ❌ Not running"
    fi
    
    if sudo systemctl is-active --quiet docker; then
        echo "   Docker Service: ✅ Running"
    else
        echo "   Docker Service: ❌ Not running"
    fi
    
    echo ""
    print_success "Installation script completed!"
    echo "For detailed documentation, see README.md in the project directory."
}

# Run main function
main "$@" 