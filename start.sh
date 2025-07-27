#!/bin/bash

echo "🚀 Starting 0G Galileo Prometheus Monitoring System"

# Dependency check
echo "🔍 Checking dependencies..."

# Docker check
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed."
    echo "   Please install Docker using:"
    echo "   sudo apt update && sudo apt install -y docker.io"
    exit 1
else
    echo "✅ Docker installed: $(docker --version)"
fi

# Docker Compose check
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed."
    echo "   Please install Docker Compose using:"
    echo "   sudo apt install -y docker-compose"
    exit 1
else
    echo "✅ Docker Compose installed: $(docker-compose --version)"
fi

# Start Docker service
echo "🔧 Starting Docker service..."
systemctl start docker 2>/dev/null || true
systemctl enable docker 2>/dev/null || true

# Stop existing running services
echo "🛑 Stopping existing running services..."
echo "   - Stopping Docker Compose services"
docker-compose down 2>/dev/null || true

echo "   - Stopping existing processes"
pkill -f "unified-metrics" 2>/dev/null || true
pkill -f "prometheus" 2>/dev/null || true
pkill -f "grafana" 2>/dev/null || true
pkill -f "nginx" 2>/dev/null || true

# Check and stop processes using required ports
echo "   - Stopping processes using ports 8080, 3000, 9090, 80"
for port in 8080 3000 9090 80; do
    pid=$(lsof -ti:$port 2>/dev/null)
    if [ ! -z "$pid" ]; then
        echo "     Stopping process $pid using port $port"
        kill -9 $pid 2>/dev/null || true
    fi
done

# Additional Docker cleanup
echo "   - Cleaning up Docker resources"
docker stop $(docker ps -q --filter "name=og-galileo") 2>/dev/null || true
docker rm $(docker ps -aq --filter "name=og-galileo") 2>/dev/null || true
docker network prune -f 2>/dev/null || true

# Wait a moment
sleep 3

# Check Node Exporter status
echo "🔍 Checking Node Exporter status..."
if pgrep -x "node_exporter" > /dev/null; then
    echo "✅ Node Exporter is already running (port 9200)"
else
    echo "⚠️  Node Exporter is not running."
    echo "   Please start Node Exporter using:"
    echo "   node_exporter --web.listen-address=:9200 &"
    echo ""
    read -p "Start Node Exporter now? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📦 Starting Node Exporter..."
        node_exporter --web.listen-address=:9200 &
        sleep 3
        if pgrep -x "node_exporter" > /dev/null; then
            echo "✅ Node Exporter started successfully."
        else
            echo "❌ Failed to start Node Exporter."
        fi
    fi
fi

# Start services with Docker Compose
echo "📦 Starting Docker Compose services..."
if docker-compose up -d; then
    echo "✅ Docker Compose services started successfully."
else
    echo "❌ Failed to start Docker Compose services."
    echo "   Attempting cleanup and retry..."
    
    # Clean up and retry
    docker-compose down 2>/dev/null || true
    docker network prune -f 2>/dev/null || true
    sleep 5
    
    echo "   Retrying Docker Compose start..."
    if docker-compose up -d; then
        echo "✅ Docker Compose services started successfully on retry."
    else
        echo "❌ Failed to start Docker Compose services after retry."
        echo "   Check logs with: docker-compose logs"
        echo "   Check port conflicts with: lsof -i :8080 -i :3000 -i :9090 -i :80"
    exit 1
    fi
fi

# Wait for services to start
echo "⏳ Waiting for services to start..."
sleep 30

# Check service status
echo "🔍 Checking service status..."
docker-compose ps

# Service status summary
echo ""
echo "📊 Service Status Summary:"
echo "========================"
for service in unified-metrics prometheus grafana nginx; do
    status=$(docker-compose ps -q $service 2>/dev/null)
    if [ ! -z "$status" ]; then
        echo "✅ $service: Running"
    else
        echo "❌ $service: Stopped"
    fi
done

echo ""
echo "✅ 0G Galileo Prometheus Monitoring System started!"
echo ""
echo "📊 Access Information:"
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
echo "🔍 Check Metrics:"
echo "   Unified Metrics: curl http://localhost:8080/all-metrics | grep og_galileo"
echo "   Beacon Chain:    curl http://localhost:8080/all-metrics | grep beacon_block_signed"
echo "   System:          curl http://localhost:8080/all-metrics | grep node_cpu"
echo ""
echo "📋 View Logs:"
echo "   docker-compose logs -f unified-metrics"
echo "   docker-compose logs -f nginx"
echo ""
echo "⚠️  Beacon Chain Characteristics:"
echo "   - Block signing status is determined by previous block's LastCommit info"
echo "   - Current block N signing status = Block N-1 signing information"
echo ""
echo "🔧 Firewall Configuration (for external access):"
echo "   sudo ufw allow 80/tcp"
echo "   sudo ufw allow 3000/tcp"
echo "   sudo ufw allow 8080/tcp"
echo "   sudo ufw allow 9090/tcp"
echo "   sudo ufw allow 9200/tcp"
echo ""
echo "🛑 To stop: docker-compose down" 