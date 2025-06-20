#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

# Load environment variables from .env file
if [ -f .env ]; then
  export $(cat .env | sed 's/#.*//g' | xargs)
fi

echo ">>> Creating Resource Group: $RESOURCE_GROUP..."
az group create --name $RESOURCE_GROUP --location $LOCATION

echo ">>> Creating Virtual Machine: $VM_NAME..."
VM_INFO=$(az vm create \
  --resource-group $RESOURCE_GROUP \
  --name $VM_NAME \
  --image UbuntuLTS \
  --size Standard_B1s \
  --admin-username $VM_USERNAME \
  --generate-ssh-keys)

# Extract and save the public IP using jq
PUBLIC_IP=$(echo $VM_INFO | jq -r '.publicIpAddress')
echo $PUBLIC_IP > .azure_vm_ip
echo ">>> VM created with Public IP: $PUBLIC_IP"

echo ">>> Opening port 80 (HTTP)..."
az vm open-port --port 80 --resource-group $RESOURCE_GROUP --name $VM_NAME --priority 100

echo "✅ Provisioning complete."