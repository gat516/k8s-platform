#!/bin/bash
set -e

INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

echo "🔧 Installing tools to $INSTALL_DIR"

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

install_go() {
    if command -v go &> /dev/null; then
        echo "✅ go already installed ($(go version))"
        return
    fi

    echo "📦 Installing Go..."
    GO_VERSION=$(curl -fsSL "https://go.dev/dl/?mode=json" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['version'])")
    curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz

    if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin
    echo "✅ go installed (${GO_VERSION})"
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
    install_go
    install_kubectl
    install_helm
    add_to_path

    echo ""
    echo "🎉 Tools installed!"
    echo "📍 User-space tools in: $INSTALL_DIR"
    echo "📍 Go installed in: /usr/local/go"
    echo ""
    echo "Still needed (system-level):"
    echo "  • Docker: sudo dnf install docker"
    echo "  • k3s: installed by setup-local.sh"
}

main