# 0G Galileo Prometheus Monitoring System

A comprehensive **integrated Prometheus monitoring system** for 0G Chain Galileo validators and system resources.

## 🚀 Key Features

- **Unified Metrics Collector**: Provides validator, beacon chain, and system metrics from a single port
- **Beacon Chain Optimization**: Accurate signing status tracking using -1 block query method
- **Real-time Monitoring**: Block status updates every 10 seconds
- **External Access Support**: External connectivity through nginx reverse proxy
- **Grafana Dashboards**: Complete visualization including State Timeline panels

## 📁 Project Structure

```
0g_prometheus/
├── unified-metrics/          # Unified metrics collector (Go)
│   ├── main.go              # Main application
│   ├── go.mod               # Go module configuration
│   └── Dockerfile           # Docker build configuration
├── grafana/                 # Grafana configuration
│   ├── provisioning/        # Auto-configuration
│   │   ├── datasources/     # Data source configuration
│   │   └── dashboards/      # Dashboard configuration
│   └── dashboards/          # Dashboard JSON files
├── nginx/                   # nginx reverse proxy
│   └── nginx.conf           # nginx configuration
├── docker-compose.yml       # Docker Compose configuration
├── prometheus.yml           # Prometheus configuration
├── prometheus-unified.yml   # Unified Prometheus configuration
├── install.sh               # Installation script
├── start.sh                 # Start script
└── README.md               # Project documentation
```

## 🏗️ Architecture

### Components

1. **unified-metrics** (Port 8080)
   - Validator metrics collection
   - Beacon chain block signing status tracking
   - System metrics integration
   - Mempool status estimation

2. **Prometheus** (Port 9090)
   - Metrics storage and querying
   - Unified data source management

3. **Grafana** (Port 3000)
   - Dashboards and visualization
   - State Timeline panels

4. **nginx** (Port 80)
   - Reverse proxy
   - External access support

5. **Node Exporter** (Port 9200)
   - System resource metrics

### Data Flow

```
0G Galileo Node → unified-metrics → Prometheus → Grafana
Node Exporter   ↗                                    ↓
0G Node        ↗                              nginx (External Access)
```

## 📊 Collected Metrics

### Validator Metrics
- `og_galileo_validator_block_height` - Current block height
- `og_galileo_validator_active_set` - Number of active validators
- `og_galileo_validator_is_bonded` - Validator bonding status
- `og_galileo_validator_missed_blocks` - Number of missed blocks
- `og_galileo_validator_beacon_block_signed` - **Beacon chain block signing status** ⭐
- `og_galileo_validator_mempool_size` - Mempool size (estimated)

### System Metrics
- `node_cpu_seconds_total` - CPU usage
- `node_memory_MemTotal_bytes` - Memory usage
- `node_filesystem_size_bytes` - Disk usage
- `node_network_receive_bytes_total` - Network receive traffic

### 0G Node Metrics
- `beacon_kit_state_block_tx_gas_used` - Gas usage per block
- `beacon_kit_*` - Beacon chain node status

## ⚠️ Beacon Chain Characteristics

0G Galileo is a **beacon chain**, where block signing status is determined using **-1 block previous signing information**.

```
Current block N signing status = Block N-1 LastCommit signing information
```

## 🚀 Installation and Setup

### Quick Installation

```bash
# Clone the repository
git clone <repository-url> 0g_prometheus
cd 0g_prometheus

# Run the installation script
chmod +x install.sh
./install.sh
```

### Manual Installation

#### 1. Install Dependencies
```bash
# Update system packages
sudo apt update

# Install required packages
sudo apt install -y curl wget git docker.io docker-compose

# Start and enable Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group (optional)
sudo usermod -aG docker $USER
```

#### 2. Clone Project
```bash
cd /root
git clone <repository-url> 0g_prometheus
cd 0g_prometheus
```

#### 3. Setup Node Exporter
```bash
# Check if Node Exporter is running
ps aux | grep node_exporter

# If not running, start it
node_exporter --web.listen-address=:9200 &
```

#### 4. Start Services
```bash
# Auto start (recommended)
./start.sh

# Or manual start
docker-compose up -d
```

## 🌐 Access Information

### Direct Access
- **Unified Metrics**: http://localhost:8080/all-metrics
- **Grafana**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Node Exporter**: http://localhost:9200

### nginx Reverse Proxy
- **Unified Metrics**: http://localhost/all-metrics/
- **Grafana**: http://localhost/grafana/
- **Prometheus**: http://localhost/prometheus/
- **Node Exporter**: http://localhost/node-exporter/

### External Access
- **Unified Metrics**: http://YOUR_SERVER_IP/all-metrics/
- **Grafana**: http://YOUR_SERVER_IP/grafana/
- **Prometheus**: http://YOUR_SERVER_IP/prometheus/

## 📋 Dashboards

### 1. 0G Galileo Beacon Chain Validator Monitoring
Basic validator metrics dashboard

### 2. 0G Galileo Beacon Chain Block Signing Timeline ⭐
**State Timeline** visualization for block-by-block signing status
- Beacon chain characteristics reflected (-1 block basis)
- Real-time updates (every 5 seconds)
- Signing rate tracking

### 3. 0G Galileo System Monitoring
System resource monitoring dashboard
- CPU, memory, disk usage
- Network traffic
- System load

## 🔧 Configuration

### RPC Endpoints
- **RPC**: `http://57.129.73.24:50657`
- **Prometheus**: `http://57.129.73.24:50660`

### Tracked Validators
- `0x1188d8FF55D1af13147f08178347B0E0fD569831` (validator1)
- `0x21f5C524FCA565dD50841fF4b92A7220Aa5B0BDD` (validator2)

## 🛠️ Development

### Build Unified Metrics Collector
```bash
cd unified-metrics
go mod tidy
go build -o main .
```

### Local Testing
```bash
cd unified-metrics
go run main.go
```

### Check Metrics
```bash
# Check unified metrics
curl http://localhost:8080/all-metrics | grep og_galileo

# Check beacon chain metrics
curl http://localhost:8080/all-metrics | grep beacon_block_signed

# Check system metrics
curl http://localhost:8080/all-metrics | grep node_cpu
```

## 🔍 Troubleshooting

### 1. RPC Connection Issues
```bash
# Check RPC endpoint
curl http://57.129.73.24:50657/status
```

### 2. Metrics Collection Issues
```bash
# Check logs
docker-compose logs unified-metrics

# Check metrics endpoint
curl http://localhost:8080/all-metrics
```

### 3. Grafana Data Source Issues
- Grafana → Settings → Data Sources
- Select "Unified Prometheus (All Metrics)"
- URL: `http://unified-metrics:8080/all-metrics`

### 4. External Access Issues
```bash
# Firewall configuration
sudo ufw allow 80/tcp
sudo ufw allow 3000/tcp
sudo ufw allow 8080/tcp
sudo ufw allow 9090/tcp
sudo ufw allow 9200/tcp
```

## 📈 Performance Optimization

- **Memory Management**: Keep only recent 1000 block information
- **Duplicate Prevention**: Avoid reprocessing already processed blocks
- **Error Handling**: Log previous block query failures
- **nginx Caching**: Performance improvement through static resource caching

## 🔒 Security Considerations

- **Default Password Change**: Recommended to change Grafana default password
- **Firewall Configuration**: Open only necessary ports
- **SSL Certificate**: Recommended to use HTTPS in production environment
- **Access Control**: Consider IP whitelist configuration

## 📄 License

MIT License 

## 🤝 Contributing

This project is an open-source project for the 0G Chain Galileo community.

---

**Developer**: 0G Galileo Monitoring Team  
**Version**: 1.0.0  
**Last Updated**: 2025-07-23 