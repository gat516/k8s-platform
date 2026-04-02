#!/bin/bash
set -e

echo "🧹 Cleaning up local k8s-platform development environment..."

confirm() {
    read -p "Are you sure you want to remove the local k3s cluster? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Cleanup cancelled"
        exit 1
    fi
}

remove_k3s() {
    echo "🗑️  Removing k3s..."
    
    if command -v k3s-uninstall.sh &> /dev/null; then
        sudo k3s-uninstall.sh
        echo "✅ k3s removed"
    else
        echo "⚠️  k3s-uninstall.sh not found, k3s may not be installed"
    fi
}

cleanup_kubeconfig() {
    echo "🧹 Cleaning up kubeconfig..."
    
    if [ -f ~/.kube/config ]; then
        rm ~/.kube/config
        echo "✅ kubeconfig removed"
    fi
}

main() {
    confirm
    remove_k3s
    cleanup_kubeconfig
    
    echo ""
    echo "🎉 Cleanup complete! Local environment has been removed."
    echo ""
}

main