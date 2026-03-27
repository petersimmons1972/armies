FROM cgr.dev/chainguard/python:3.11

# Install git and openssh-client for sync functionality
# Chainguard Python is Alpine-based; use apk
USER root
RUN apk add --no-cache git openssh-client
USER nonroot

# Set working directory
WORKDIR /workspace

# Copy project files
# Chainguard runs as UID 65532 (nonroot) by default — use that user/group
COPY --chown=nonroot:nonroot . .

# Install armies package
RUN pip install -e . --no-cache-dir

# Profile data lives in a volume mount — never inside the image
VOLUME ["/home/nonroot/.armies"]

ENV HOME=/home/nonroot

ENTRYPOINT ["armies"]
CMD ["--help"]
