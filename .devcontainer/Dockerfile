# Use the official multi-platform Node.js image
FROM node:20

# Arguments for versions and configuration, allowing overrides during build
ARG TZ=Etc/UTC
ARG TARGETARCH
ARG USERNAME=node
ARG GOLANG_VERSION=1.23.10
ARG GOLANGCI_LINT_VERSION=v1.55.2
ARG GIT_DELTA_VERSION=0.18.2

# Set environment variables
ENV TZ="$TZ"
ENV SHELL=/bin/zsh
ENV DEVCONTAINER=true
# Go environment
ENV GOLANG_VERSION=$GOLANG_VERSION
ENV GOPATH=/home/node/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:/usr/local/share/npm-global/bin:$PATH
# NPM global environment
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global

# --- System-level installations as root ---

# 1. Install all dependencies in a single layer to optimize image size.
RUN apt-get update && apt-get install -y --no-install-recommends \
  less git procps sudo fzf zsh man-db unzip gnupg2 curl wget \
  build-essential gh iptables ipset iproute2 dnsutils aggregate jq redis-server \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# 2. Install Go for the correct architecture dynamically
RUN wget "https://go.dev/dl/go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz" -O go.tar.gz \
  && tar -C /usr/local -xzf go.tar.gz \
  && rm go.tar.gz

# 3. Install other architecture-aware tools (golangci-lint, git-delta)
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}
RUN wget "https://github.com/dandavison/delta/releases/download/${GIT_DELTA_VERSION}/git-delta_${GIT_DELTA_VERSION}_${TARGETARCH}.deb" \
  && dpkg -i "git-delta_${GIT_DELTA_VERSION}_${TARGETARCH}.deb" \
  && rm "git-delta_${GIT_DELTA_VERSION}_${TARGETARCH}.deb"

# 4. Create directories and set up permissions
RUN mkdir -p /workspace /commandhistory /home/$USERNAME/.claude $NPM_CONFIG_PREFIX $GOPATH/bin \
  && chown -R $USERNAME:$USERNAME /workspace /commandhistory /home/$USERNAME $NPM_CONFIG_PREFIX $GOPATH

# 5. Set up persisted command history
RUN touch /commandhistory/.bash_history \
  && chown $USERNAME:$USERNAME /commandhistory/.bash_history

# 6. Copy scripts and set up sudo permissions for the node user
COPY init-firewall.sh start-redis.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/init-firewall.sh /usr/local/bin/start-redis.sh \
  && echo "$USERNAME ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh" > /etc/sudoers.d/node-firewall \
  && echo "$USERNAME ALL=(root) NOPASSWD: /usr/sbin/service" >> /etc/sudoers.d/node-firewall \
  && chmod 0440 /etc/sudoers.d/node-firewall

# --- Switch to non-root user ---
USER $USERNAME
WORKDIR /workspace

# 7. Install user-specific tools (Go, NPM)
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN npm install -g @anthropic-ai/claude-code

# 8. Set up Zsh with a default theme and plugins
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.0/zsh-in-docker.sh)" -- \
  -p git \
  -p fzf \
  -a "source /usr/share/doc/fzf/examples/key-bindings.zsh" \
  -a "source /usr/share/doc/fzf/examples/completion.zsh" \
  -a "export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  -x