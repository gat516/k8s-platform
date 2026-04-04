#!/usr/bin/env bash
set -euo pipefail

# get-kubeconfig.sh — retrieve kubeconfig from a freshly provisioned k3s VPS,
# rewrite the server address, and print the base64-encoded value for use as
# the KUBECONFIG_B64 GitHub Actions secret.
#
# Usage:
#   ./infrastructure/get-kubeconfig.sh <VPS_IP>
#   VPS_IP=1.2.3.4 ./infrastructure/get-kubeconfig.sh

VPS_IP="${1:-${VPS_IP:-}}"

if [[ -z "$VPS_IP" ]]; then
  echo "Error: VPS_IP is required." >&2
  echo "Usage: $0 <VPS_IP>  or  VPS_IP=1.2.3.4 $0" >&2
  exit 1
fi

KUBECONFIG_PATH="${HOME}/.kube/k3s-platform.yaml"

# Ensure ~/.kube exists
mkdir -p "${HOME}/.kube"

echo "==> Fetching kubeconfig from root@${VPS_IP}..."
# k3s writes a world-readable kubeconfig at this path (--write-kubeconfig-mode 644)
ssh -o StrictHostKeyChecking=accept-new \
    -o BatchMode=yes \
    "root@${VPS_IP}" \
    "cat /etc/rancher/k3s/k3s.yaml" > "${KUBECONFIG_PATH}"

# k3s always writes 127.0.0.1 as the server address; replace with the actual VPS IP
echo "==> Rewriting server address (127.0.0.1 -> ${VPS_IP})..."
sed -i "s/127\.0\.0\.1/${VPS_IP}/g" "${KUBECONFIG_PATH}"

# Restrict permissions — kubeconfig contains a cluster admin credential
chmod 600 "${KUBECONFIG_PATH}"

echo "==> Kubeconfig saved to ${KUBECONFIG_PATH}"
echo ""
echo "==> Verify connectivity:"
echo "    kubectl --kubeconfig=${KUBECONFIG_PATH} get nodes"
echo ""
echo "==> Base64-encoded value for the KUBECONFIG_B64 GitHub Actions secret:"
echo "    (copy everything on the line below)"
echo ""

# base64 -w0 is Linux-specific (no line wrapping); macOS uses base64 without -w0
if base64 --version 2>&1 | grep -q GNU; then
  base64 -w0 "${KUBECONFIG_PATH}"
else
  # macOS / BSD base64
  base64 "${KUBECONFIG_PATH}"
fi

echo ""
echo "==> Add the secret (paste the base64 value printed above):"
echo "    gh secret set KUBECONFIG_B64 --body \"<paste-value-above>\""
