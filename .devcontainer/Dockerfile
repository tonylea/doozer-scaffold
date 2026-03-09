FROM mcr.microsoft.com/devcontainers/base:ubuntu

# Base packages (tmux + procps are required by overstory)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    ca-certificates \
    tmux \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Install Bun (required runtime for overstory)
# Install as vscode user since devcontainer runs as non-root
USER vscode
RUN curl -fsSL https://bun.sh/install | bash
ENV BUN_INSTALL="/home/vscode/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"
