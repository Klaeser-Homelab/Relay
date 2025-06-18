# Relay Proxmox Implementation Checklist

**Version:** 1.0  
**Based on:** specs/plan.md  
**Implementation Phase:** Phase 1 - Foundation and Integration  
**Estimated Time:** 4-6 weeks

---

## Pre-Implementation Planning

### ☐ **Hardware Assessment and Procurement**

#### Minimum Hardware Requirements (Development Environment)
- [ ] **CPU**: 16 cores (Intel Xeon or AMD EPYC)
- [ ] **RAM**: 64GB DDR4-3200 ECC
- [ ] **Storage Primary**: 1TB NVMe SSD (ZFS compatible)
- [ ] **Storage Backup**: 4TB HDD for snapshots
- [ ] **Network**: 1Gbps uplink, 10Gbps internal capability
- [ ] **Redundancy**: Consider dual PSU for production use

#### Recommended Hardware (Production Environment)
- [ ] **CPU**: 32 cores with hyper-threading
- [ ] **RAM**: 128GB DDR4-3200 or DDR5 ECC
- [ ] **Storage Primary**: 2TB NVMe SSD in RAID-1
- [ ] **Storage Backup**: 8TB HDD array for backups
- [ ] **Network**: 10Gbps uplink with 25Gbps internal
- [ ] **Redundancy**: Full redundancy (PSU, networking, storage)

### ☐ **Network Planning**

#### VLAN Configuration
- [ ] **VLAN 100**: Development Network (Relay containers)
- [ ] **VLAN 200**: Services Network (Redis, databases, monitoring)
- [ ] **VLAN 300**: Management Network (Proxmox, backups)
- [ ] **VLAN 400**: External Access (Traefik gateway, VPN)

#### IP Address Allocation
```bash
# Document your IP ranges
Development Network (VLAN 100): 10.100.0.0/24
Services Network (VLAN 200):    10.200.0.0/24  
Management Network (VLAN 300):  10.300.0.0/24
External Access (VLAN 400):     10.400.0.0/24
```

#### Required External Services
- [ ] **GitHub API Access**: Ensure firewall allows GitHub API endpoints
- [ ] **OpenAI API Access**: Verify OpenAI Realtime API connectivity
- [ ] **Claude API Access**: Confirm Anthropic API endpoint access
- [ ] **Package Repositories**: Allow access to Docker Hub, npm, Go modules
- [ ] **Certificate Authority**: Access for Let's Encrypt or internal CA

### ☐ **Security Prerequisites**

#### SSL/TLS Certificates
- [ ] **Domain Name**: Register or configure internal domain for services
- [ ] **Wildcard Certificate**: Obtain SSL cert for *.yourdomain.com
- [ ] **Internal CA**: Setup internal Certificate Authority (optional)

#### Authentication Systems
- [ ] **GitHub Personal Access Tokens**: Create tokens for repository access
- [ ] **OpenAI API Keys**: Obtain API keys with Realtime API access
- [ ] **Claude API Keys**: Get Anthropic API credentials
- [ ] **SSH Key Pairs**: Generate SSH keys for Git operations

#### VPN Configuration (Optional but Recommended)
- [ ] **WireGuard Server**: Plan VPN server deployment
- [ ] **Client Configurations**: Prepare VPN profiles for developers
- [ ] **Access Policies**: Define which services require VPN access

---

## Phase 1: Proxmox Infrastructure Setup

### ☐ **Proxmox VE Installation**

#### Basic Installation
```bash
# Download Proxmox VE ISO
# Boot from USB/ISO and follow installation wizard

# Post-installation commands (run on Proxmox host)
# 1. Update system
apt update && apt upgrade -y

# 2. Configure repositories (remove enterprise repo if no subscription)
echo "deb http://download.proxmox.com/debian/pve bookworm pve-no-subscription" > /etc/apt/sources.list.d/pve-no-subscription.list
```

- [ ] **ISO Download**: Download latest Proxmox VE ISO
- [ ] **Installation**: Complete base installation with ZFS storage
- [ ] **Network Configuration**: Configure management network interface
- [ ] **Updates**: Apply all available updates
- [ ] **Repository Configuration**: Configure community repositories
- [ ] **Web Interface Access**: Verify access to https://proxmox-ip:8006

#### Storage Configuration
```bash
# Create ZFS storage pools
zpool create -f relay-storage /dev/nvme0n1 /dev/nvme1n1  # RAID-1 for primary
zpool create -f relay-backup /dev/sda /dev/sdb           # RAID-1 for backups

# Configure Proxmox storage
pvesm add zfspool relay-storage --pool relay-storage --content vztmpl,rootdir,images
pvesm add zfspool relay-backup --pool relay-backup --content backup,snippets
```

- [ ] **Primary Storage**: Configure NVMe SSD storage pool
- [ ] **Backup Storage**: Configure HDD backup pool  
- [ ] **Proxmox Storage**: Add storage pools to Proxmox configuration
- [ ] **Test Storage**: Verify storage pools are healthy and accessible

### ☐ **Network Configuration**

#### Bridge and VLAN Setup
```bash
# Edit /etc/network/interfaces
# Example configuration:

auto lo
iface lo inet loopback

auto eno1
iface eno1 inet manual

auto vmbr0
iface vmbr0 inet static
    address 10.300.0.10/24
    gateway 10.300.0.1
    bridge-ports eno1
    bridge-stp off
    bridge-fd 0

# VLAN interfaces
auto vmbr0.100
iface vmbr0.100 inet static
    address 10.100.0.1/24
    vlan-raw-device vmbr0

auto vmbr0.200  
iface vmbr0.200 inet static
    address 10.200.0.1/24
    vlan-raw-device vmbr0

auto vmbr0.400
iface vmbr0.400 inet static
    address 10.400.0.1/24
    vlan-raw-device vmbr0
```

- [ ] **Management Bridge**: Configure primary bridge (vmbr0)
- [ ] **VLAN Configuration**: Setup VLANs for different networks
- [ ] **Network Restart**: Apply network configuration
- [ ] **Connectivity Test**: Verify all VLANs are accessible

#### Firewall Configuration
```bash
# Basic Proxmox firewall rules
# Configure via Proxmox web interface or CLI

# Allow SSH access
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Allow Proxmox web interface
iptables -A INPUT -p tcp --dport 8006 -j ACCEPT

# Allow VNC/SPICE for container access
iptables -A INPUT -p tcp --dport 5900:5999 -j ACCEPT
```

- [ ] **Basic Firewall**: Configure initial firewall rules
- [ ] **SSH Access**: Ensure SSH access to Proxmox host
- [ ] **Web Interface**: Secure web interface access
- [ ] **Container Access**: Allow VNC/console access to containers

---

## Phase 2: Container Environment Preparation

### ☐ **LXC Container Creation**

#### Create Base Development Container
```bash
# Download container template
pveam update
pveam download local ubuntu-22.04-standard_22.04-1_amd64.tar.zst

# Create LXC container
pct create 100 \
    local:vztmpl/ubuntu-22.04-standard_22.04-1_amd64.tar.zst \
    --arch amd64 \
    --cores 4 \
    --memory 8192 \
    --swap 2048 \
    --net0 name=eth0,bridge=vmbr0,tag=100,ip=10.100.0.100/24,gw=10.100.0.1 \
    --storage relay-storage \
    --rootfs relay-storage:50 \
    --hostname relay-dev-primary \
    --password changeme123 \
    --unprivileged 1 \
    --features nesting=1,keyctl=1
```

- [ ] **Template Download**: Download Ubuntu 22.04 LXC template
- [ ] **Container Creation**: Create primary development container
- [ ] **Network Assignment**: Assign to development VLAN
- [ ] **Resource Allocation**: Configure CPU, memory, and storage
- [ ] **Container Start**: Start container and verify connectivity

#### Container Post-Configuration
```bash
# Enter container
pct enter 100

# Update system
apt update && apt upgrade -y

# Install essential packages
apt install -y curl wget git build-essential software-properties-common

# Configure for Docker
echo "lxc.apparmor.profile: unconfined" >> /etc/pve/lxc/100.conf
echo "lxc.cgroup2.devices.allow: a" >> /etc/pve/lxc/100.conf
echo "lxc.cap.drop:" >> /etc/pve/lxc/100.conf
echo "lxc.mount.auto: proc:rw sys:rw" >> /etc/pve/lxc/100.conf

# Restart container
pct reboot 100
```

- [ ] **System Updates**: Update container to latest packages
- [ ] **Essential Tools**: Install development prerequisites
- [ ] **Docker Preparation**: Configure container for Docker support
- [ ] **Container Restart**: Apply configuration changes

### ☐ **Docker Installation and Configuration**

#### Install Docker in Container
```bash
# Enter the container
pct enter 100

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Add user to docker group
usermod -aG docker root

# Test Docker installation
docker run hello-world
```

- [ ] **Docker Installation**: Install Docker CE in container
- [ ] **Docker Compose**: Install Docker Compose plugin
- [ ] **Service Configuration**: Enable Docker service
- [ ] **Test Installation**: Verify Docker is working correctly

#### Create Docker Networks
```bash
# Create custom networks for Relay services
docker network create --driver bridge relay-development
docker network create --driver bridge relay-services
docker network create --driver bridge relay-monitoring

# List networks to verify
docker network ls
```

- [ ] **Development Network**: Create network for Relay containers
- [ ] **Services Network**: Create network for shared services
- [ ] **Monitoring Network**: Create network for monitoring stack
- [ ] **Network Verification**: Verify all networks are created

### ☐ **Persistent Storage Setup**

#### Create Volume Directories
```bash
# Create directory structure for persistent storage
mkdir -p /opt/relay/{repositories,config,data,logs,backups}
mkdir -p /opt/relay/data/{redis,postgres,prometheus,grafana}
mkdir -p /opt/relay/config/{traefik,monitoring,relay}

# Set appropriate permissions
chown -R 1000:1000 /opt/relay
chmod -R 755 /opt/relay

# Create Docker volumes
docker volume create relay-repositories
docker volume create relay-config
docker volume create relay-redis-data
docker volume create relay-postgres-data
docker volume create relay-monitoring-data
```

- [ ] **Directory Structure**: Create base directory structure
- [ ] **Permissions**: Set appropriate ownership and permissions
- [ ] **Docker Volumes**: Create named volumes for persistence
- [ ] **Backup Preparation**: Ensure backup directories are accessible

---

## Phase 3: Core Service Deployment

### ☐ **Traefik Gateway Configuration**

#### Create Traefik Configuration
```yaml
# Create /opt/relay/config/traefik/traefik.yml
api:
  dashboard: true
  debug: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
  file:
    filename: /etc/traefik/dynamic.yml

certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@yourdomain.com
      storage: acme.json
      httpChallenge:
        entryPoint: web

# Global redirect to https
http:
  redirections:
    entryPoint:
      to: websecure
      scheme: https
```

```yaml
# Create /opt/relay/config/traefik/dynamic.yml
http:
  middlewares:
    default-headers:
      headers:
        frameDeny: true
        sslRedirect: true
        browserXssFilter: true
        contentTypeNosniff: true
        forceSTSHeader: true
        stsIncludeSubdomains: true
        stsPreload: true
        stsSeconds: 31536000

tls:
  options:
    default:
      sslProtocols:
        - "TLSv1.2"
        - "TLSv1.3"
      minVersion: "VersionTLS12"
      cipherSuites:
        - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
        - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
        - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
```

- [ ] **Traefik Configuration**: Create main Traefik configuration
- [ ] **SSL Configuration**: Configure Let's Encrypt or custom certificates
- [ ] **Security Headers**: Configure security middleware
- [ ] **Dynamic Configuration**: Setup file-based provider for routing

#### Deploy Traefik Container
```yaml
# Create /opt/relay/docker-compose.yml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    container_name: relay-traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /opt/relay/config/traefik:/etc/traefik:ro
      - /opt/relay/data/traefik:/data
    networks:
      - relay-services
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.dashboard.rule=Host(`traefik.relay.local`)"
      - "traefik.http.routers.dashboard.tls=true"
      - "traefik.http.services.dashboard.loadbalancer.server.port=8080"

networks:
  relay-services:
    external: true
```

- [ ] **Docker Compose**: Create Traefik service definition
- [ ] **Port Configuration**: Configure HTTP, HTTPS, and dashboard ports
- [ ] **Volume Mounts**: Mount configuration and data directories
- [ ] **Service Start**: Start Traefik and verify dashboard access

### ☐ **Redis Cache and Job Queue**

#### Deploy Redis Service
```yaml
# Add to docker-compose.yml
  redis:
    image: redis:7-alpine
    container_name: relay-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - relay-redis-data:/data
      - /opt/relay/config/redis.conf:/usr/local/etc/redis/redis.conf:ro
    command: redis-server /usr/local/etc/redis/redis.conf
    networks:
      - relay-services
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
```

```conf
# Create /opt/relay/config/redis.conf
bind 0.0.0.0
port 6379
timeout 0
tcp-keepalive 300
daemonize no
supervised systemd
pidfile /var/run/redis_6379.pid
loglevel notice
logfile ""
databases 16
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data
maxmemory 512mb
maxmemory-policy allkeys-lru
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
```

- [ ] **Redis Configuration**: Create Redis configuration file
- [ ] **Service Definition**: Add Redis to Docker Compose
- [ ] **Data Persistence**: Configure Redis data persistence
- [ ] **Health Check**: Verify Redis is running and accessible

### ☐ **PostgreSQL Database (Optional)**

#### Deploy PostgreSQL for Shared Data
```yaml
# Add to docker-compose.yml (if needed for future features)
  postgres:
    image: postgres:15-alpine
    container_name: relay-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: relay
      POSTGRES_USER: relay
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - relay-postgres-data:/var/lib/postgresql/data
      - /opt/relay/config/postgres/init:/docker-entrypoint-initdb.d:ro
    networks:
      - relay-services
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U relay"]
      interval: 30s
      timeout: 10s
      retries: 3
```

```sql
-- Create /opt/relay/config/postgres/init/init.sql
-- Initial database setup
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Voice sessions table
CREATE TABLE voice_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    session_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Repository metadata table  
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    git_url TEXT NOT NULL,
    local_path TEXT,
    dependencies JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

- [ ] **PostgreSQL Service**: Add PostgreSQL to stack (optional)
- [ ] **Database Initialization**: Create initial schema
- [ ] **Environment Variables**: Configure database credentials
- [ ] **Connection Test**: Verify database connectivity

---

## Phase 4: Relay Application Deployment

### ☐ **Relay Development Container Build**

#### Create Dockerfile for Relay Dev Environment
```dockerfile
# Create /opt/relay/dockerfiles/relay-dev-env/Dockerfile
FROM ubuntu:22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV NODE_VERSION=20
ENV GO_VERSION=1.21.5
ENV CLAUDE_CODE_VERSION=latest
ENV RELAY_USER=relay
ENV RELAY_HOME=/home/relay
ENV WORKSPACE=/workspace

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl wget git build-essential \
    python3 python3-pip python3-venv \
    docker.io docker-compose \
    redis-tools postgresql-client \
    vim nano tmux htop jq \
    openssh-client ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs

# Install Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Install Claude Code CLI
RUN npm install -g @anthropic/claude-code

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
    gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
    tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update && apt-get install -y gh

# Create relay user
RUN useradd -m -s /bin/bash -u 1000 ${RELAY_USER} \
    && usermod -aG docker ${RELAY_USER} \
    && echo "${RELAY_USER} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Set up Go environment
ENV PATH="/usr/local/go/bin:${RELAY_HOME}/.local/bin:${RELAY_HOME}/go/bin:${PATH}"
ENV GOPATH="${RELAY_HOME}/go"

# Switch to relay user
USER ${RELAY_USER}
WORKDIR ${RELAY_HOME}

# Create directory structure
RUN mkdir -p ${RELAY_HOME}/.config/relay \
    && mkdir -p ${RELAY_HOME}/go/{bin,src,pkg} \
    && mkdir -p ${WORKSPACE}

# Install development tools
RUN go install github.com/air-verse/air@latest \
    && npm install -g yarn pnpm typescript ts-node nodemon

# Configure git defaults
RUN git config --global init.defaultBranch main \
    && git config --global pull.rebase false \
    && git config --global user.name "Relay Developer" \
    && git config --global user.email "developer@relay.local"

# Copy Relay components (will be mounted as volumes in production)
COPY --chown=${RELAY_USER}:${RELAY_USER} entrypoint.sh ${RELAY_HOME}/
RUN chmod +x ${RELAY_HOME}/entrypoint.sh

# Expose ports
EXPOSE 8080 3000 5173

# Set working directory
WORKDIR ${WORKSPACE}

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/home/relay/entrypoint.sh"]
CMD ["relay", "start", "--voice-enabled"]
```

```bash
# Create /opt/relay/dockerfiles/relay-dev-env/entrypoint.sh
#!/bin/bash

set -e

echo "Starting Relay Development Environment..."

# Setup environment variables
export RELAY_HOME="/home/relay"
export WORKSPACE="/workspace"
export PATH="/usr/local/go/bin:${RELAY_HOME}/.local/bin:${RELAY_HOME}/go/bin:${PATH}"

# Configure Git if credentials are provided
if [ -n "$GIT_USER_NAME" ]; then
    git config --global user.name "$GIT_USER_NAME"
fi

if [ -n "$GIT_USER_EMAIL" ]; then
    git config --global user.email "$GIT_USER_EMAIL"
fi

# Setup GitHub CLI if token is provided
if [ -n "$GITHUB_TOKEN" ]; then
    echo "$GITHUB_TOKEN" | gh auth login --with-token
fi

# Setup Claude Code if API key is provided
if [ -n "$CLAUDE_API_KEY" ]; then
    claude auth login --api-key "$CLAUDE_API_KEY"
fi

# Initialize workspace if empty
if [ ! -f "$WORKSPACE/.relay-initialized" ]; then
    echo "Initializing workspace..."
    cd "$WORKSPACE"
    
    # Clone Relay repository if URL provided
    if [ -n "$RELAY_REPO_URL" ]; then
        git clone "$RELAY_REPO_URL" relay
        cd relay
    fi
    
    touch "$WORKSPACE/.relay-initialized"
fi

# Start services based on command
if [ "$1" = "relay" ] && [ "$2" = "start" ]; then
    echo "Starting Relay services..."
    
    # Start voice server in background
    if [ -d "$WORKSPACE/relay/voice-server-js" ]; then
        cd "$WORKSPACE/relay/voice-server-js"
        npm install
        npm run dev &
    fi
    
    # Start frontend in background  
    if [ -d "$WORKSPACE/relay/voice-frontend" ]; then
        cd "$WORKSPACE/relay/voice-frontend"
        npm install
        npm run dev &
    fi
    
    # Keep container running
    tail -f /dev/null
else
    # Execute provided command
    exec "$@"
fi
```

- [ ] **Dockerfile Creation**: Create comprehensive development container
- [ ] **Entrypoint Script**: Create initialization and startup script
- [ ] **Build Image**: Build the relay-dev-env Docker image
- [ ] **Test Container**: Verify container starts and runs correctly

#### Build and Test Development Container
```bash
# Build the development container
cd /opt/relay/dockerfiles/relay-dev-env
docker build -t relay-dev-env:latest .

# Test the container
docker run --rm -it \
    -e GITHUB_TOKEN=$GITHUB_TOKEN \
    -e CLAUDE_API_KEY=$CLAUDE_API_KEY \
    -e GIT_USER_NAME="Your Name" \
    -e GIT_USER_EMAIL="your@email.com" \
    -v /opt/relay/repositories:/workspace \
    relay-dev-env:latest bash

# Inside container, verify installations
node --version
go version
gh --version
claude --version
```

- [ ] **Image Build**: Build development container image
- [ ] **Test Run**: Run container with test parameters
- [ ] **Verify Tools**: Check all development tools are installed
- [ ] **Volume Mounts**: Test workspace volume mounting

### ☐ **Repository Manager Service**

#### Create Repository Manager
```go
// Create /opt/relay/src/repo-manager/main.go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/gorilla/mux"
    "github.com/go-redis/redis/v8"
)

type RepoManager struct {
    config *Config
    redis  *redis.Client
}

type Config struct {
    Repositories []Repository `yaml:"repositories"`
    WorkspaceDir string       `yaml:"workspace_dir"`
    RedisURL     string       `yaml:"redis_url"`
}

type Repository struct {
    Name         string   `yaml:"name"`
    URL          string   `yaml:"url"`
    Dependencies []string `yaml:"dependencies"`
    AutoSync     bool     `yaml:"auto_sync"`
}

func (rm *RepoManager) syncRepository(repoName string) error {
    repo := rm.findRepository(repoName)
    if repo == nil {
        return fmt.Errorf("repository %s not found", repoName)
    }

    repoPath := filepath.Join(rm.config.WorkspaceDir, repo.Name)
    
    // Clone if doesn't exist
    if _, err := os.Stat(repoPath); os.IsNotExist(err) {
        cmd := exec.Command("git", "clone", repo.URL, repoPath)
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("failed to clone %s: %v", repo.Name, err)
        }
        log.Printf("Cloned repository %s", repo.Name)
    }

    // Pull latest changes
    cmd := exec.Command("git", "-C", repoPath, "pull", "origin", "main")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to pull %s: %v", repo.Name, err)
    }
    
    log.Printf("Synced repository %s", repo.Name)
    return nil
}

func (rm *RepoManager) findRepository(name string) *Repository {
    for _, repo := range rm.config.Repositories {
        if repo.Name == name {
            return &repo
        }
    }
    return nil
}

func (rm *RepoManager) handleSync(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    repoName := vars["repo"]
    
    if err := rm.syncRepository(repoName); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "message": fmt.Sprintf("Repository %s synced", repoName),
    })
}

func (rm *RepoManager) handleSyncAll(w http.ResponseWriter, r *http.Request) {
    var results []map[string]interface{}
    
    for _, repo := range rm.config.Repositories {
        result := map[string]interface{}{
            "repository": repo.Name,
        }
        
        if err := rm.syncRepository(repo.Name); err != nil {
            result["status"] = "error"
            result["error"] = err.Error()
        } else {
            result["status"] = "success"
        }
        
        results = append(results, result)
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "completed",
        "results": results,
    })
}

func main() {
    repoManager := &RepoManager{
        config: &Config{
            WorkspaceDir: "/workspace",
            RedisURL:     "redis://redis:6379",
        },
    }

    // Initialize Redis client
    opt, _ := redis.ParseURL(repoManager.config.RedisURL)
    repoManager.redis = redis.NewClient(opt)

    // Setup HTTP routes
    r := mux.NewRouter()
    r.HandleFunc("/sync/{repo}", repoManager.handleSync).Methods("POST")
    r.HandleFunc("/sync-all", repoManager.handleSyncAll).Methods("POST")
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
    }).Methods("GET")

    log.Println("Repository Manager starting on :9090")
    log.Fatal(http.ListenAndServe(":9090", r))
}
```

```dockerfile
# Create /opt/relay/dockerfiles/repo-manager/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o repo-manager .

FROM alpine:latest
RUN apk --no-cache add ca-certificates git openssh-client
WORKDIR /root/

COPY --from=builder /app/repo-manager .

EXPOSE 9090
CMD ["./repo-manager"]
```

```yaml
# Add to docker-compose.yml
  repo-manager:
    build: ./dockerfiles/repo-manager
    container_name: relay-repo-manager
    restart: unless-stopped
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - REDIS_URL=redis://redis:6379
      - WORKSPACE_DIR=/workspace
    volumes:
      - /opt/relay/repositories:/workspace
      - /opt/relay/config/repo-config.yaml:/config.yaml:ro
    ports:
      - "9090:9090"
    networks:
      - relay-services
    depends_on:
      - redis
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

- [ ] **Repository Manager Code**: Create Go-based repository manager
- [ ] **Dockerfile**: Create container for repo manager
- [ ] **Service Integration**: Add to Docker Compose stack
- [ ] **API Testing**: Test repository synchronization endpoints

### ☐ **Voice Server Integration**

#### Deploy Enhanced Voice Server
```yaml
# Add to docker-compose.yml
  voice-server:
    build: 
      context: /opt/relay/repositories/relay/voice-server-js
      dockerfile: Dockerfile
    container_name: relay-voice-server
    restart: unless-stopped
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}  
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
      - REDIS_URL=redis://redis:6379
      - NODE_ENV=production
    volumes:
      - /opt/relay/repositories:/workspace:ro
    ports:
      - "8080:8080"
    networks:
      - relay-services
      - relay-development
    depends_on:
      - redis
      - repo-manager
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.voice-server.rule=Host(`voice.relay.local`)"
      - "traefik.http.routers.voice-server.tls=true"
      - "traefik.http.services.voice-server.loadbalancer.server.port=8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

- [ ] **Voice Server Build**: Build voice server from existing code
- [ ] **Environment Configuration**: Set up API keys and Redis connection
- [ ] **Traefik Integration**: Configure routing through Traefik
- [ ] **Health Checks**: Verify voice server is responding

#### Deploy Voice Frontend
```yaml
# Add to docker-compose.yml
  voice-frontend:
    build:
      context: /opt/relay/repositories/relay/voice-frontend
      dockerfile: Dockerfile.prod
    container_name: relay-voice-frontend
    restart: unless-stopped
    environment:
      - VITE_API_URL=https://voice.relay.local
      - VITE_WS_URL=wss://voice.relay.local
    networks:
      - relay-development
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`relay.local`)"
      - "traefik.http.routers.frontend.tls=true"
      - "traefik.http.services.frontend.loadbalancer.server.port=80"
    depends_on:
      - voice-server
```

```dockerfile
# Create /opt/relay/repositories/relay/voice-frontend/Dockerfile.prod
FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Frontend Build**: Create production frontend build
- [ ] **Nginx Configuration**: Configure web server for frontend
- [ ] **Routing Setup**: Configure Traefik routing for frontend
- [ ] **Integration Test**: Verify frontend connects to voice server

---

## Phase 5: Monitoring and Observability

### ☐ **Prometheus Monitoring Stack**

#### Deploy Prometheus
```yaml
# Add to docker-compose.yml
  prometheus:
    image: prom/prometheus:latest
    container_name: relay-prometheus
    restart: unless-stopped
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    volumes:
      - /opt/relay/config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - relay-monitoring-data:/prometheus
    ports:
      - "9090:9090"
    networks:
      - relay-services
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.prometheus.rule=Host(`prometheus.relay.local`)"
      - "traefik.http.routers.prometheus.tls=true"
```

```yaml
# Create /opt/relay/config/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'relay-voice-server'
    static_configs:
      - targets: ['voice-server:8080']
    metrics_path: '/metrics'

  - job_name: 'relay-repo-manager'
    static_configs:
      - targets: ['repo-manager:9090']
    metrics_path: '/metrics'

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']

  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
```

- [ ] **Prometheus Configuration**: Create monitoring configuration
- [ ] **Service Discovery**: Configure service monitoring targets
- [ ] **Data Retention**: Set appropriate data retention policies
- [ ] **Web Interface**: Verify Prometheus web interface access

#### Deploy Grafana
```yaml
# Add to docker-compose.yml
  grafana:
    image: grafana/grafana:latest
    container_name: relay-grafana
    restart: unless-stopped
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
      - GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource
    volumes:
      - relay-monitoring-data:/var/lib/grafana
      - /opt/relay/config/grafana/provisioning:/etc/grafana/provisioning:ro
    ports:
      - "3000:3000"
    networks:
      - relay-services
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.grafana.rule=Host(`grafana.relay.local`)"
      - "traefik.http.routers.grafana.tls=true"
    depends_on:
      - prometheus
```

```yaml
# Create /opt/relay/config/grafana/provisioning/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

- [ ] **Grafana Deployment**: Deploy Grafana with Prometheus integration
- [ ] **Data Source Configuration**: Configure Prometheus as data source
- [ ] **Dashboard Import**: Import relevant dashboards for monitoring
- [ ] **Alerting Setup**: Configure alerting rules and notifications

### ☐ **Log Aggregation with Loki**

#### Deploy Loki and Promtail
```yaml
# Add to docker-compose.yml
  loki:
    image: grafana/loki:latest
    container_name: relay-loki
    restart: unless-stopped
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - /opt/relay/config/loki:/etc/loki:ro
      - /opt/relay/data/loki:/loki
    ports:
      - "3100:3100"
    networks:
      - relay-services

  promtail:
    image: grafana/promtail:latest
    container_name: relay-promtail
    restart: unless-stopped
    command: -config.file=/etc/promtail/config.yml
    volumes:
      - /opt/relay/config/promtail:/etc/promtail:ro
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    networks:
      - relay-services
    depends_on:
      - loki
```

```yaml
# Create /opt/relay/config/loki/local-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    address: 127.0.0.1
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
    final_sleep: 0s
  chunk_idle_period: 5m
  chunk_retain_period: 30s

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb:
    directory: /loki/index
  filesystem:
    directory: /loki/chunks

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: false
  retention_period: 0s
```

- [ ] **Loki Configuration**: Setup log aggregation server
- [ ] **Promtail Configuration**: Configure log collection agent
- [ ] **Log Sources**: Configure container and system log collection
- [ ] **Grafana Integration**: Add Loki as data source in Grafana

---

## Phase 6: Security and Access Control

### ☐ **SSL/TLS Certificate Management**

#### Configure Let's Encrypt (Production)
```bash
# For production with real domain
# Ensure DNS points to your server

# Create Let's Encrypt configuration
mkdir -p /opt/relay/config/traefik/acme
chmod 600 /opt/relay/config/traefik/acme

# Update traefik.yml to use Let's Encrypt
# This is automatically handled by Traefik configuration above
```

#### Self-Signed Certificates (Development)
```bash
# For development/internal use
mkdir -p /opt/relay/config/ssl

# Generate CA private key
openssl genrsa -out /opt/relay/config/ssl/ca-key.pem 4096

# Generate CA certificate
openssl req -new -x509 -days 365 -key /opt/relay/config/ssl/ca-key.pem \
    -sha256 -out /opt/relay/config/ssl/ca.pem -subj "/C=US/ST=CA/L=SF/O=Relay/CN=Relay CA"

# Generate server private key
openssl genrsa -out /opt/relay/config/ssl/server-key.pem 4096

# Generate server certificate signing request
openssl req -subj "/C=US/ST=CA/L=SF/O=Relay/CN=*.relay.local" -sha256 \
    -new -key /opt/relay/config/ssl/server-key.pem -out /opt/relay/config/ssl/server.csr

# Generate server certificate
openssl x509 -req -days 365 -sha256 -in /opt/relay/config/ssl/server.csr \
    -CA /opt/relay/config/ssl/ca.pem -CAkey /opt/relay/config/ssl/ca-key.pem \
    -out /opt/relay/config/ssl/server-cert.pem -CAcreateserial
```

- [ ] **Certificate Generation**: Create SSL certificates for development
- [ ] **Let's Encrypt Setup**: Configure automatic certificate renewal
- [ ] **Traefik Integration**: Configure Traefik to use certificates
- [ ] **Certificate Validation**: Verify SSL certificates are working

### ☐ **Authentication and Authorization**

#### Configure Environment Variables
```bash
# Create /opt/relay/.env file
# NEVER commit this file to version control

# API Keys
OPENAI_API_KEY=sk-your-openai-api-key
CLAUDE_API_KEY=your-claude-api-key
GITHUB_TOKEN=ghp_your-github-token
GEMINI_API_KEY=your-gemini-api-key

# Database passwords
POSTGRES_PASSWORD=secure-random-password
REDIS_PASSWORD=another-secure-password
GRAFANA_PASSWORD=grafana-admin-password

# JWT secrets
JWT_SECRET=very-long-random-string-for-jwt-signing

# Webhook secrets
WEBHOOK_SECRET=secret-for-github-webhooks

# User configuration
GIT_USER_NAME="Your Name"
GIT_USER_EMAIL="your@email.com"

# Repository configuration
RELAY_REPO_URL=https://github.com/yourusername/relay.git
```

- [ ] **Environment File**: Create secure environment configuration
- [ ] **API Key Management**: Configure all required API keys
- [ ] **Password Generation**: Generate secure passwords for services
- [ ] **Git Configuration**: Set up Git user configuration

#### Setup Access Control
```yaml
# Create /opt/relay/config/auth/rbac.yml
roles:
  developer:
    permissions:
      - "container:access"
      - "repository:read,write"
      - "voice:command"
      - "monitoring:read"
    resources:
      cpu_limit: "4"
      memory_limit: "8Gi"
      
  team_lead:
    permissions:
      - "container:access,manage"
      - "repository:read,write,admin"
      - "voice:command,configure"
      - "monitoring:read,write"
    resources:
      cpu_limit: "8"
      memory_limit: "16Gi"
      
  admin:
    permissions:
      - "*"
    resources:
      cpu_limit: "unlimited"
      memory_limit: "unlimited"

users:
  - username: "developer1"
    role: "developer"
    email: "dev1@company.com"
  - username: "teamlead1"
    role: "team_lead"
    email: "lead1@company.com"
```

- [ ] **RBAC Configuration**: Define roles and permissions
- [ ] **User Management**: Configure user accounts and roles
- [ ] **Resource Limits**: Set resource quotas per role
- [ ] **Access Validation**: Test access control mechanisms

---

## Phase 7: Testing and Validation

### ☐ **Infrastructure Testing**

#### Container Health Verification
```bash
# Test all containers are running
docker-compose ps

# Check container health
docker-compose exec voice-server curl -f http://localhost:8080/health
docker-compose exec repo-manager wget -qO- http://localhost:9090/health
docker-compose exec redis redis-cli ping

# Test network connectivity
docker-compose exec voice-server ping redis
docker-compose exec repo-manager ping voice-server
```

- [ ] **Container Status**: Verify all containers are running
- [ ] **Health Checks**: Confirm all health checks pass
- [ ] **Network Connectivity**: Test inter-container communication
- [ ] **Volume Mounts**: Verify persistent storage is working

#### Service Integration Testing
```bash
# Test Traefik routing
curl -H "Host: relay.local" https://10.100.0.1/
curl -H "Host: voice.relay.local" https://10.100.0.1/health
curl -H "Host: grafana.relay.local" https://10.100.0.1/

# Test repository synchronization
curl -X POST http://10.100.0.1:9090/sync-all

# Test Redis connectivity
redis-cli -h 10.100.0.1 -p 6379 ping
```

- [ ] **Traefik Routing**: Test reverse proxy functionality
- [ ] **Service APIs**: Verify all service endpoints respond
- [ ] **Database Connectivity**: Test Redis and PostgreSQL connections
- [ ] **SSL Certificates**: Verify SSL/TLS is working correctly

### ☐ **Voice Command Testing**

#### Basic Voice Functionality
```javascript
// Test voice server WebSocket connection
const WebSocket = require('ws');

const ws = new WebSocket('wss://voice.relay.local');

ws.on('open', function open() {
    console.log('Connected to voice server');
    
    // Test voice command
    ws.send(JSON.stringify({
        type: 'test_command',
        data: {
            command: 'list repositories'
        }
    }));
});

ws.on('message', function message(data) {
    console.log('Received:', JSON.parse(data));
});
```

- [ ] **WebSocket Connection**: Test voice server WebSocket connectivity
- [ ] **Basic Commands**: Test simple voice commands
- [ ] **Error Handling**: Verify error responses are proper
- [ ] **Audio Processing**: Test voice recording and processing

#### Claude Code Integration Testing
```bash
# Test Claude Code CLI installation
docker-compose exec voice-server claude --version

# Test API connectivity
docker-compose exec voice-server claude auth status

# Test basic code analysis
docker-compose exec voice-server claude analyze /workspace/relay/README.md
```

- [ ] **Claude Code CLI**: Verify installation and authentication
- [ ] **API Connectivity**: Test connection to Claude API
- [ ] **Basic Operations**: Test code analysis functionality
- [ ] **Error Handling**: Verify proper error responses

### ☐ **Multi-Repository Testing**

#### Repository Synchronization
```bash
# Test repository cloning
curl -X POST http://voice.relay.local/api/repositories/sync \
    -H "Content-Type: application/json" \
    -d '{"repositories": ["relay", "test-repo"]}'

# Verify repositories exist
ls -la /opt/relay/repositories/

# Test dependency resolution
curl -X GET http://voice.relay.local/api/repositories/dependencies/relay
```

- [ ] **Repository Cloning**: Test automatic repository synchronization
- [ ] **Dependency Management**: Verify dependency graph handling
- [ ] **Update Synchronization**: Test repository update propagation
- [ ] **Conflict Resolution**: Test handling of merge conflicts

#### Cross-Repository Operations
```bash
# Test cross-repository analysis
curl -X POST http://voice.relay.local/api/analyze \
    -H "Content-Type: application/json" \
    -d '{"repositories": ["frontend", "backend"], "type": "dependencies"}'

# Test coordinated changes
curl -X POST http://voice.relay.local/api/implement \
    -H "Content-Type: application/json" \
    -d '{"feature": "add authentication", "repositories": ["frontend", "backend"]}'
```

- [ ] **Cross-Repository Analysis**: Test multi-repo code analysis
- [ ] **Coordinated Changes**: Test synchronized modifications
- [ ] **Impact Analysis**: Verify cross-project impact detection
- [ ] **Integration Testing**: Test end-to-end multi-repo workflows

---

## Phase 8: Production Deployment

### ☐ **Environment Preparation**

#### Production Configuration
```bash
# Copy production environment template
cp /opt/relay/.env.example /opt/relay/.env.production

# Configure production values
# Edit .env.production with production API keys and configurations

# Create production docker-compose override
cat > /opt/relay/docker-compose.prod.yml << EOF
version: '3.8'

services:
  voice-server:
    environment:
      - NODE_ENV=production
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G

  repo-manager:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 1G

  redis:
    command: redis-server --requirepass \${REDIS_PASSWORD}
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 1G
        reservations:
          cpus: '0.25'
          memory: 512M
EOF
```

- [ ] **Production Environment**: Configure production-specific settings
- [ ] **Resource Limits**: Set appropriate resource constraints
- [ ] **Security Configuration**: Enable authentication where applicable
- [ ] **Performance Tuning**: Optimize configuration for production load

#### Backup and Recovery Setup
```bash
# Create backup scripts
mkdir -p /opt/relay/scripts/backup

cat > /opt/relay/scripts/backup/backup-daily.sh << 'EOF'
#!/bin/bash

set -e

BACKUP_DIR="/opt/relay/backups/$(date +%Y-%m-%d)"
mkdir -p "$BACKUP_DIR"

# Backup configuration
tar -czf "$BACKUP_DIR/config-backup.tar.gz" /opt/relay/config/

# Backup repositories
tar -czf "$BACKUP_DIR/repositories-backup.tar.gz" /opt/relay/repositories/

# Backup container volumes
docker run --rm -v relay-redis-data:/data -v "$BACKUP_DIR:/backup" alpine \
    tar -czf /backup/redis-data.tar.gz -C /data .

docker run --rm -v relay-postgres-data:/data -v "$BACKUP_DIR:/backup" alpine \
    tar -czf /backup/postgres-data.tar.gz -C /data .

# Backup monitoring data
docker run --rm -v relay-monitoring-data:/data -v "$BACKUP_DIR:/backup" alpine \
    tar -czf /backup/monitoring-data.tar.gz -C /data .

# Remove backups older than 30 days
find /opt/relay/backups/ -type d -mtime +30 -exec rm -rf {} +

echo "Backup completed: $BACKUP_DIR"
EOF

chmod +x /opt/relay/scripts/backup/backup-daily.sh

# Setup cron job for daily backups
echo "0 2 * * * /opt/relay/scripts/backup/backup-daily.sh" | crontab -
```

- [ ] **Backup Scripts**: Create automated backup procedures
- [ ] **Cron Jobs**: Schedule regular backups
- [ ] **Recovery Testing**: Test backup restoration procedures
- [ ] **Offsite Storage**: Configure offsite backup storage

### ☐ **Migration from Existing Setup**

#### Data Migration
```bash
# Create migration script
cat > /opt/relay/scripts/migrate-existing.sh << 'EOF'
#!/bin/bash

set -e

echo "Starting migration from existing Relay setup..."

# Backup existing configuration
if [ -d "$HOME/.config/relay" ]; then
    echo "Backing up existing Relay configuration..."
    cp -r "$HOME/.config/relay" /opt/relay/config/relay-existing-backup/
fi

# Export existing projects
if command -v relay &> /dev/null; then
    echo "Exporting existing Relay projects..."
    relay export --format json > /tmp/relay-projects-export.json
fi

# Migrate repositories
if [ -d "$HOME/Code" ]; then
    echo "Migrating existing repositories..."
    rsync -av "$HOME/Code/" /opt/relay/repositories/
fi

# Import configuration into new setup
echo "Importing configuration into containerized setup..."
docker-compose exec voice-server relay import --config /tmp/relay-projects-export.json

echo "Migration completed successfully!"
EOF

chmod +x /opt/relay/scripts/migrate-existing.sh
```

- [ ] **Migration Script**: Create automated migration procedure
- [ ] **Configuration Backup**: Backup existing Relay configuration
- [ ] **Repository Migration**: Move existing repositories to container storage
- [ ] **Validation**: Verify migrated data integrity

#### User Training and Documentation
```bash
# Create user documentation
mkdir -p /opt/relay/docs

cat > /opt/relay/docs/USER_GUIDE.md << 'EOF'
# Relay Containerized Setup - User Guide

## Accessing Your Development Environment

### Web Interface
- **Frontend**: https://relay.local
- **Voice Server**: https://voice.relay.local
- **Monitoring**: https://grafana.relay.local
- **Traefik Dashboard**: https://traefik.relay.local

### Development Container Access
```bash
# SSH into development container
docker-compose exec voice-server bash

# Access specific repository
docker-compose exec voice-server bash -c "cd /workspace/relay && bash"
```

### Voice Commands
- "List my repositories"
- "Switch to [repository-name]"
- "Analyze the code structure"
- "Implement [feature-description]"
- "Run tests for [service-name]"

### Troubleshooting
- Check container status: `docker-compose ps`
- View logs: `docker-compose logs [service-name]`
- Restart service: `docker-compose restart [service-name]`
EOF
```

- [ ] **User Documentation**: Create comprehensive user guides
- [ ] **Training Sessions**: Conduct user training sessions
- [ ] **Quick Reference**: Provide voice command reference cards
- [ ] **Support Procedures**: Establish support and troubleshooting procedures

### ☐ **Go-Live Procedures**

#### Final Pre-Launch Checklist
```bash
# Final verification script
cat > /opt/relay/scripts/pre-launch-check.sh << 'EOF'
#!/bin/bash

echo "=== Pre-Launch Verification ==="

# Check all containers are running
echo "Checking container status..."
docker-compose ps

# Test all health endpoints
echo "Testing health endpoints..."
curl -f https://voice.relay.local/health || echo "❌ Voice server health check failed"
curl -f http://repo-manager:9090/health || echo "❌ Repo manager health check failed"

# Test voice functionality
echo "Testing voice server WebSocket..."
# Add WebSocket test here

# Test repository access
echo "Testing repository access..."
ls -la /opt/relay/repositories/ || echo "❌ Repository access failed"

# Test monitoring
echo "Testing monitoring stack..."
curl -f https://grafana.relay.local/api/health || echo "❌ Grafana health check failed"

# Test SSL certificates
echo "Testing SSL certificates..."
openssl s_client -connect relay.local:443 -verify_return_error < /dev/null || echo "❌ SSL certificate invalid"

echo "=== Pre-Launch Check Completed ==="
EOF

chmod +x /opt/relay/scripts/pre-launch-check.sh
```

- [ ] **Pre-Launch Testing**: Run comprehensive pre-launch tests
- [ ] **Performance Validation**: Verify system performance under load
- [ ] **Security Audit**: Complete security review and penetration testing
- [ ] **Rollback Plan**: Prepare rollback procedures if issues arise

#### Launch and Monitoring
```bash
# Production deployment
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Monitor initial deployment
watch docker-compose ps

# Monitor logs for errors
docker-compose logs -f

# Monitor resource usage
docker stats
```

- [ ] **Production Deployment**: Deploy to production environment
- [ ] **Real-Time Monitoring**: Monitor system performance and logs
- [ ] **User Acceptance Testing**: Conduct final user acceptance testing
- [ ] **Performance Monitoring**: Verify system meets performance targets

---

## Post-Deployment Optimization

### ☐ **Performance Tuning**

#### Resource Optimization
- [ ] **CPU Usage Analysis**: Monitor and optimize CPU usage patterns
- [ ] **Memory Optimization**: Tune memory allocation and garbage collection
- [ ] **Storage Performance**: Optimize storage I/O and caching strategies
- [ ] **Network Optimization**: Fine-tune network configuration and bandwidth

#### Scaling Preparation
- [ ] **Load Testing**: Conduct comprehensive load testing
- [ ] **Horizontal Scaling**: Plan for multiple container instances
- [ ] **Database Scaling**: Prepare for database scaling requirements
- [ ] **Monitoring Alerts**: Configure proactive monitoring and alerting

### ☐ **Security Hardening**

#### Security Audit
- [ ] **Vulnerability Scanning**: Regular automated vulnerability scans
- [ ] **Access Review**: Review and audit user access permissions
- [ ] **Certificate Management**: Automate certificate renewal and monitoring
- [ ] **Backup Security**: Encrypt backups and test restoration procedures

#### Compliance
- [ ] **Audit Logging**: Ensure comprehensive audit trail
- [ ] **Data Protection**: Implement data protection and privacy measures
- [ ] **Compliance Reporting**: Generate compliance reports for security audits
- [ ] **Incident Response**: Establish incident response procedures

### ☐ **Continuous Improvement**

#### Monitoring and Analytics
- [ ] **Usage Analytics**: Track voice command usage patterns
- [ ] **Performance Metrics**: Monitor system performance trends
- [ ] **User Feedback**: Collect and analyze user feedback
- [ ] **Feature Usage**: Track adoption of new features

#### Future Planning
- [ ] **Phase 2 Planning**: Plan for multi-repository management features
- [ ] **Team Expansion**: Prepare for team growth and collaboration features
- [ ] **Technology Updates**: Plan for technology stack updates
- [ ] **Feature Roadmap**: Develop long-term feature development roadmap

---

## Support and Maintenance

### ☐ **Maintenance Procedures**

#### Regular Maintenance Tasks
```bash
# Weekly maintenance script
cat > /opt/relay/scripts/weekly-maintenance.sh << 'EOF'
#!/bin/bash

echo "Starting weekly maintenance..."

# Update container images
docker-compose pull

# Clean up unused Docker resources
docker system prune -f

# Update system packages
apt update && apt upgrade -y

# Check disk space
df -h

# Rotate logs
logrotate /etc/logrotate.conf

# Test backups
/opt/relay/scripts/backup/test-restore.sh

echo "Weekly maintenance completed."
EOF
```

- [ ] **Update Procedures**: Establish regular update procedures
- [ ] **Log Rotation**: Configure log rotation and cleanup
- [ ] **Resource Cleanup**: Regular cleanup of unused resources
- [ ] **Health Monitoring**: Continuous health monitoring and alerting

#### Troubleshooting Guide
- [ ] **Common Issues**: Document common issues and solutions
- [ ] **Error Codes**: Create error code reference guide
- [ ] **Support Contacts**: Establish support escalation procedures
- [ ] **Knowledge Base**: Maintain comprehensive troubleshooting knowledge base

---

## Conclusion

This implementation checklist provides a comprehensive guide for deploying the Relay voice-controlled development platform on Proxmox infrastructure. The checklist is organized into phases that build upon each other, ensuring a systematic and reliable deployment process.

### Key Success Factors

1. **Infrastructure Foundation**: Proper Proxmox setup with adequate resources and networking
2. **Container Orchestration**: Well-configured Docker environment with proper networking and storage
3. **Service Integration**: Seamless integration between all Relay components
4. **Security Implementation**: Comprehensive security measures from the start
5. **Monitoring and Observability**: Proper monitoring and alerting for proactive management
6. **Testing and Validation**: Thorough testing at each phase to ensure reliability
7. **Documentation and Training**: Comprehensive documentation and user training

### Timeline Estimate

- **Phase 1-2**: Infrastructure Setup (1-2 weeks)
- **Phase 3-4**: Service Deployment (2-3 weeks)  
- **Phase 5-6**: Monitoring and Security (1 week)
- **Phase 7**: Testing and Validation (1 week)
- **Phase 8**: Production Deployment (1 week)

**Total Estimated Time**: 6-8 weeks for complete implementation

### Next Steps

After completing this checklist, you'll have a fully functional Relay development platform running on Proxmox infrastructure, ready for Phase 2 implementation focusing on advanced multi-repository management and AI-powered development capabilities.

---

*This checklist should be customized based on your specific infrastructure requirements, security policies, and organizational needs. Regular updates to this document should reflect lessons learned and best practices discovered during implementation.*