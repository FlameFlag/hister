FROM golang:1.24

# Switch workdir do build directory
WORKDIR /hister/build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build

# Switch workdir to final destination
WORKDIR /hister

# Copy binary, remove build dir
RUN cp ./build/hister .
RUN rm -rf ./build

# Install required utilities for user/group management
RUN apt-get update && apt-get install -y --no-install-recommends \
    gosu \
    && rm -rf /var/lib/apt/lists/*


#VOLUME $HOME/.config/hister/

EXPOSE 4433

CMD ["/hister/hister", "listen"]
