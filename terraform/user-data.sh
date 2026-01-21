#!/bin/bash
set -e

# Logging
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "========================================="
echo "Starting k3s installation"
echo "========================================="

# Update system
yum update -y

# Install required packages
yum install -y \
    curl \
    wget \
    git \
    vim \
    htop \
    jq \
    net-tools

# Install Docker (for building images if needed)
yum install -y docker
systemctl enable docker
systemctl start docker
usermod -aG docker ec2-user

# Install k3s WITHOUT Traefik and WITHOUT Flannel
# We'll use Calico instead of Flannel
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="${k3s_version}" sh -s - \
    --disable traefik \
    --flannel-backend=none \
    --disable-network-policy \
    --write-kubeconfig-mode 644

# Wait for k3s to be ready
echo "Waiting for k3s to be ready..."
sleep 30

# Install Calico
echo "Installing Calico CNI..."
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.27.0/manifests/tigera-operator.yaml

# Wait for operator
sleep 10

# Create Calico installation
cat <<EOF | kubectl apply -f -
apiVersion: operator.tigera.io/v1
kind: Installation
metadata:
  name: default
spec:
  calicoNetwork:
    ipPools:
    - blockSize: 26
      cidr: 10.42.0.0/16
      encapsulation: VXLANCrossSubnet
      natOutgoing: Enabled
      nodeSelector: all()
EOF

# Wait for Calico to be ready
echo "Waiting for Calico to be ready..."
sleep 60

# Install kubectl completion
kubectl completion bash > /etc/bash_completion.d/kubectl

# Configure kubeconfig for ec2-user
mkdir -p /home/ec2-user/.kube
cp /etc/rancher/k3s/k3s.yaml /home/ec2-user/.kube/config
chown -R ec2-user:ec2-user /home/ec2-user/.kube
chmod 600 /home/ec2-user/.kube/config

# Add kubectl alias
echo "alias k=kubectl" >> /home/ec2-user/.bashrc
echo "complete -F __start_kubectl k" >> /home/ec2-user/.bashrc

# Create directory for k8s manifests
mkdir -p /home/ec2-user/k8s
chown ec2-user:ec2-user /home/ec2-user/k8s

# Install helm (for kube-prometheus-stack)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Add helm repos for later use
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Create a marker file to indicate completion
echo "$(date): k3s installation completed successfully" > /home/ec2-user/k3s-ready

# Print status
echo "========================================="
echo "k3s installation completed!"
echo "========================================="
kubectl get nodes
kubectl get pods -A

# Set hostname
hostnamectl set-hostname payment-gateway-k3s

echo "User data script completed at $(date)" >> /var/log/user-data-complete