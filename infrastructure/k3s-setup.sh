#!/usr/bin/env bash
set -euo pipefail

# k3s-setup.sh — bootstrap a fresh Ubuntu 22.04 VPS for k8s-platform
# Run as root on the VPS after SSH key auth is configured.

echo "==> Checking for existing k3s installation..."
if command -v k3s &>/dev/null; then
  echo "k3s already installed, skipping."
else
  echo "==> Installing k3s..."
  curl -sfL https://get.k3s.io | sh -s - \
    --disable=servicelb \
    --write-kubeconfig-mode 644
  echo "==> Waiting for node to be ready..."
  until k3s kubectl get nodes | grep -q " Ready"; do sleep 2; done
  echo "k3s is ready."
fi

echo "==> Configuring UFW..."
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp   # SSH
ufw allow 80/tcp   # HTTP (Traefik)
ufw allow 443/tcp  # HTTPS (Traefik)
ufw allow 6443/tcp # Kubernetes API
ufw --force enable
echo "UFW configured."

echo ""
echo "==> Done. To configure remote access:"
echo "    1. Copy the kubeconfig from the VPS:"
echo "       scp root@<VPS_IP>:/etc/rancher/k3s/k3s.yaml ~/.kube/k3s-platform.yaml"
echo "    2. Edit ~/.kube/k3s-platform.yaml: replace '127.0.0.1' with your VPS IP"
echo "    3. Test: kubectl --kubeconfig=~/.kube/k3s-platform.yaml get nodes"
echo "    4. Base64-encode and add as GitHub Actions secret KUBECONFIG_B64:"
echo "       base64 -w0 ~/.kube/k3s-platform.yaml"
