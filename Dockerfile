FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# 基本パッケージのインストール
RUN apt update && apt install -y \
    binutils \
    build-essential \
    curl \
    wget \
    ca-certificates \
    gnupg \
    sysstat \
    fonts-takao \
    fio \
    jq \
    strace \
    vim \
    && rm -rf /var/lib/apt/lists/*

# Go 1.24 のインストール
RUN wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz && \
    rm go1.24.0.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Rust のインストール
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH=$PATH:/root/.cargo/bin

# Node.js 24.x のインストール
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - && \
    apt install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

# Python 3.14 のインストール（deadsnakes PPA）
RUN apt update && apt install -y software-properties-common && \
    add-apt-repository -y ppa:deadsnakes/ppa && \
    apt update && apt install -y python3.14 && \
    update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.14 1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
