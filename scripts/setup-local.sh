#!/bin/bash
set -e

echo "🚀 Setting up local k8s-platform development environment..."

if [[ $EUID -eq 0 ]]; then
   echo "❌ This script should not be run as root"
   exit 1
fi

detect_platform() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

PLATFORM=$(detect_platform)

check_k3s() {
    if systemctl is-active --quiet k3s; then
        echo "✅ k3s is already running"
        return 0
    else
        return 1
    fi
}

install_k3s() {
    if [[ "$PLATFORM" == "macos" ]]; then
        echo "💡 On macOS, consider using k3d or Docker Desktop Kubernetes instead"
        echo "   Install k3d: brew install k3d"
        echo "   Then run: k3d cluster create k8s-platform"
        return 1
    fi
    
    if [[ "$PLATFORM" == "windows" ]]; then
        echo "💡 On Windows, use Docker Desktop Kubernetes or k3d in WSL2"
        echo "   Enable Kubernetes in Docker Desktop settings"
        return 1
    fi
    
    echo "📦 Installing k3s..."
    curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644
    
    echo "⏳ Waiting for k3s to be ready..."
    sleep 10
    
    timeout 120s bash -c 'until kubectl get nodes | grep -q Ready; do sleep 5; done'
    
    echo "✅ k3s is ready!"
}

setup_kubeconfig() {
    echo "⚙️  Setting up kubeconfig..."
    
    mkdir -p ~/.kube
    sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
    sudo chown $USER:$USER ~/.kube/config
    chmod 600 ~/.kube/config
    
    echo "✅ kubeconfig configured"
}

verify_installation() {
    echo "🔍 Verifying installation..."
    
    echo "Cluster info:"
    kubectl cluster-info
    
    echo ""
    echo "Nodes:"
    kubectl get nodes
    
    echo ""
    echo "✅ Local k3s cluster is ready!"
}

main() {
    if ! check_k3s; then
        install_k3s
    fi
    
    setup_kubeconfig
    verify_installation
    
    echo ""
    echo "🎉 Setup complete! Your local Kubernetes cluster is ready."
    echo ""
    echo "Next steps:"
    echo "  • kubectl get all --all-namespaces"
    echo "  • Continue with the roadmap items in README.md"
    echo ""
}

main