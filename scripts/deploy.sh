#!/bin/bash
set -e

# Load environment variables
if [ -f .env ]; then
  export $(cat .env | sed 's/#.*//g' | xargs)
fi

# Load the VM's IP address
if [ ! -f .azure_vm_ip ]; then
    echo "VM IP address file not found. Please run ./scripts/provision.sh first."
    exit 1
fi
PUBLIC_IP=$(cat .azure_vm_ip)

echo ">>> Deploying to VM at $PUBLIC_IP..."

# This uses a "here document" to send a block of commands to the remote server via SSH.
ssh -o StrictHostKeyChecking=no ${VM_USERNAME}@${PUBLIC_IP} << 'EOF'
  # Exit immediately if a command exits with a non-zero status.
  set -e

  # --- Check and Install Dependencies ---
  if ! command -v docker &> /dev/null; then
      echo ">>> Docker not found. Installing..."
      sudo apt-get update
      sudo apt-get install -y docker.io docker-compose git nodejs npm
      sudo usermod -aG docker $USER
      echo "Docker installed. Please re-run the deploy script after logging out and back in."
      exit 1
  fi

  # --- Clone or Update Application Code ---
  if [ ! -d "alignment" ]; then
    echo ">>> Cloning repository..."
    git clone $REPO_URL
    cd alignment
  else
    echo ">>> Updating repository..."
    cd alignment
    git pull origin main
  fi

  # --- Launch Application with Docker Compose ---
  echo ">>> Launching application with production config..."
  docker-compose -f docker-compose.prod.yml up --build -d

  echo "✅ Deployment successful."
EOF