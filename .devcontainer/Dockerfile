FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

ENV TASK_VERSION=v3.43.3 \
    GORELEASER_VERSION=v2.9.0 \
    GOLANG_CI_LINT_VERSION=v2.1.6 \
    GITHUB_CLI_VERSION=2.68.1 \
    NODE_VERSION=lts \
    SEMANTIC_RELEASE_VERSION=v24.2.3

RUN apt-get update && \
    # Determine architecture
    architecture=$(uname -m) && \
    architecture_pretty=$(dpkg --print-architecture) && \
    echo "Installing dependencies for ${architecture}" && \
    # Install networking utilities
    apt-get install -y iputils-ping dnsutils && \
    # Install Node.js
    curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - && \
    apt-get install -y nodejs && \
    # Install Task
    curl -s https://taskfile.dev/install.sh | sh -s -- -b /usr/local/bin ${TASK_VERSION} && \
    # Install GoReleaser
    curl -fsSL https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/goreleaser_Linux_${architecture_pretty}.tar.gz | tar -xzv -C /usr/local/bin goreleaser && \
    # Install golangci-lint
    curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin ${GOLANG_CI_LINT_VERSION} && \
    # Install GitHub CLI
    curl -fsSL https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_linux_${architecture_pretty}.tar.gz -o /tmp/gh_${GITHUB_CLI_VERSION}_linux_${architecture_pretty}.tar.gz && \
    tar -xzf /tmp/gh_${GITHUB_CLI_VERSION}_linux_${architecture_pretty}.tar.gz -C /tmp && \
    chmod +x /tmp/gh_${GITHUB_CLI_VERSION}_linux_${architecture_pretty}/bin/gh && \
    mv /tmp/gh_${GITHUB_CLI_VERSION}_linux_${architecture_pretty}/bin/gh /usr/local/bin/gh && \
    # Install semantic-release and plugins
    npm install -g semantic-release@${SEMANTIC_RELEASE_VERSION} \
                   @semantic-release/changelog \
                   @semantic-release/git \
                   @semantic-release/github \
                   @semantic-release/exec \
                   conventional-changelog-cli \
                   conventional-changelog-conventionalcommits && \
    # Install js-yaml for converting JSON RPC schema to YAML
    npm i -g js-yaml@4.1.0 && \
    # Clean up
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

USER vscode

RUN go install github.com/maxbrunsfeld/counterfeiter/v6@latest
RUN go install go.uber.org/mock/mockgen@v0.5.0

# Use Powerlevel10k theme
RUN git clone --depth=1 https://github.com/romkatv/powerlevel10k.git /home/vscode/.powerlevel10k
RUN echo 'source /home/vscode/.powerlevel10k/powerlevel10k.zsh-theme' >> /home/vscode/.zshrc
RUN echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> /home/vscode/.zshrc

# Shell completion
RUN echo "source <(gh completion -s zsh)" >> /home/vscode/.zshrc && \
    echo "source <(task --completion zsh)" >> /home/vscode/.zshrc && \
    echo "source <(goreleaser completion zsh)" >> /home/vscode/.zshrc
