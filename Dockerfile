# mutli-service container: Rails API + Go Processor
FROM ruby:3.4.8-slim AS base

# install system dependencies
RUN apt-get update -qq && \
    apt-get install --no-install-recommends -y \
    curl \
    ca-certificates \
    supervisor \
    wget \
    build-essential \
    libyaml-dev \
    && rm -rf /var/lib/apt/lists/*

# install Go 1.24
RUN wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz && \
    rm go1.24.0.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

WORKDIR /app

# build Go service
WORKDIR /app/go-processor

COPY go-processor/go.mod go-processor/go.sum ./
RUN go mod download

COPY go-processor/ ./
RUN go build -o processor main.go

# Setup Rails Service
WORKDIR /app/rails-api

# install rails dependencies
COPY rails-api/Gemfile rails-api/Gemfile.lock ./
RUN bundle config set --local without 'development test' && \
    bundle install

# Copy Rails application
COPY rails-api/ ./

# precompile bootsnap
RUN bundle exec bootsnap precompile --gemfile app/ lib/ || true

# Configure Supervisord
WORKDIR /app

COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# create non-root user
RUN groupadd --system --gid 1000 appuser && \
    useradd appuser --uid 1000 --gid 1000 --create-home --shell /bin/bash && \
    chown -R appuser:appuser /app

# Expose only rails port (Go is internal)
EXPOSE 3000

# Set environment
ENV RAILS_ENV=production \
    GO_SERVICE_URL=http://localhost:8080 \
    PORT=3000

# Start supervisord
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
