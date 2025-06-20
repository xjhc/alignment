#!/bin/bash
set -e

if [ -f .env ]; then
  export $(cat .env | sed 's/#.*//g' | xargs)
fi

echo "⚠️  This will permanently delete the resource group '$RESOURCE_GROUP' and all its resources."
read -p "Are you sure? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

echo ">>> Deleting Resource Group: $RESOURCE_GROUP..."
az group delete --name $RESOURCE_GROUP --yes --no-wait

# Clean up local IP file
rm -f .azure_vm_ip

echo "✅ Deletion initiated. It may take a few minutes to complete in Azure."