#!/bin/bash
set -e

echo "🔍 Checking prerequisites for kube-platform..."

MISSING_TOOLS=()

check_command() {
    if ! command -v "$1" &> /dev/null; then
        MISSING_TOOLS+=("$1")
        echo "❌ $1 is not installed"
        return 1
    else
        echo "✅ $1 is installed"
        return 0
    fi
}

echo ""
echo "Checking required tools:"
check_command "docker"
check_command "kubectl"
check_command "helm"
check_command "k3s" || echo "   💡 k3s will be installed by the setup script"

echo ""
echo "Checking optional tools:"
check_command "terraform" || echo "   💡 Required for AWS deployment"
check_command "git" || echo "   💡 Required for version control"
check_command "curl" || echo "   💡 Required for downloads"

echo ""
if [ ${#MISSING_TOOLS[@]} -eq 0 ] || [[ " ${MISSING_TOOLS[@]} " =~ " k3s " && ${#MISSING_TOOLS[@]} -eq 1 ]]; then
    echo "🎉 All required tools are available!"
    echo ""
    echo "Next steps:"
    echo "  Run: ./scripts/setup-local.sh"
else
    echo "❌ Missing required tools: ${MISSING_TOOLS[*]}"
    echo ""
    echo "Installation help:"
    echo "  Docker: https://docs.docker.com/get-docker/"
    echo "  kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/"
    echo "  Helm: https://helm.sh/docs/intro/install/"
    echo ""
    exit 1
fi