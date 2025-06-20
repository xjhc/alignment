#!/bin/bash
set -e
if [ ! -f .azure_vm_ip ]; then echo "VM IP not found. Run provision.sh first."; exit 1; fi
PUBLIC_IP=$(cat .azure_vm_ip)
VM_USERNAME=$(grep VM_USERNAME .env | cut -d '=' -f2)
ssh ${VM_USERNAME}@${PUBLIC_IP}