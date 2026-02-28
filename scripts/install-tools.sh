#!/bin/bash
set -e

INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

echo "🔧 Installing tools to $INSTALL_DIR"

install_terraform() {
    if command -v terraform &> /dev/null; then
        echo "✅ terraform already installed"
        return
    fi
    
    echo "📦 Installing terraform..."
    TF_VERSION="1.8.5"
    curl -fsSL "https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_amd64.zip" -o /tmp/terraform.zip
    cd /tmp && unzip -q terraform.zip && mv terraform "$INSTALL_DIR/"
    rm /tmp/terraform.zip
    echo "✅ terraform installed"
}

install_helm() {
    if command -v helm &> /dev/null; then
        echo "✅ helm already installed"
        return
    fi
    
    echo "📦 Installing helm..."
    curl -fsSL https://get.helm.sh/helm-v3.15.3-linux-amd64.tar.gz | tar -xz -C /tmp
    mv /tmp/linux-amd64/helm "$INSTALL_DIR/"
    rm -rf /tmp/linux-amd64
    echo "✅ helm installed"
}

install_kubectl() {
    if command -v kubectl &> /dev/null; then
        echo "✅ kubectl already installed"
        return
    fi
    
    echo "📦 Installing kubectl..."
    curl -fsSL "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" -o "$INSTALL_DIR/kubectl"
    chmod +x "$INSTALL_DIR/kubectl"
    echo "✅ kubectl installed"
}

add_to_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> ~/.bashrc
        echo "📝 Added $INSTALL_DIR to PATH in ~/.bashrc"
        echo "🔄 Run 'source ~/.bashrc' or restart your shell"
    fi
}

main() {
    install_kubectl
    install_helm
    install_terraform
    add_to_path
    
    echo ""
    echo "🎉 User-space tools installed!"
    echo "📍 Tools installed in: $INSTALL_DIR"
    echo ""
    echo "Still needed (system-level):"
    echo "  • Docker: sudo dnf install docker"
    echo "  • k3s: installed by setup-local.sh"
}

main