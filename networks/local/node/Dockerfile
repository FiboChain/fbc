FROM golang:stretch as build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y $PACKAGES

WORKDIR /fbc
# Add source files
COPY . .

# build fbc
RUN make build-linux

# Final image
FROM golang:1.17 as final

WORKDIR /fbchaind
# Copy over binaries from the build-env
COPY --from=build-env /fbc/build/fbchaind /usr/bin/fbchaind
COPY --from=build-env /fbc/build/fbchaincli /usr/bin/fbchaincli
COPY --from=build-env /fbc/networks/local/node/wrapper.sh /usr/bin/wrapper.sh

EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

