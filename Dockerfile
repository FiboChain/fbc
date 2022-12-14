# Simple usage with a mounted data directory:
# > docker build -t fbchain .
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.fbchaind:/root/.fbchaind -v ~/.fbchaincli:/root/.fbchaincli fbchain fbchaind init mynode
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.fbchaind:/root/.fbchaind -v ~/.fbchaincli:/root/.fbchaincli fbchain fbchaind start
FROM golang:alpine AS build-env

# Install minimum necessary dependencies, remove packages
RUN apk add --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev

# Set working directory for the build
WORKDIR /go/src/github.com/FiboChain/fbc

# Add source files
COPY . .

# Build fbchain
ENV GO111MODULE=on \
    GOPROXY=http://goproxy.cn
# Build Fibonacci
RUN make install

# Final image
FROM alpine:edge

WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/fbchaind /usr/bin/fbchaind
COPY --from=build-env /go/bin/fbchaincli /usr/bin/fbchaincli

# Run fbchaind by default, omit entrypoint to ease using container with fbchaincli
CMD ["fbchaind"]
