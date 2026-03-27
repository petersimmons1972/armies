FROM python:3.11-slim

# Install system dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        git \
        openssh-client \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd --create-home --shell /bin/bash armies

# Set working directory
WORKDIR /workspace

# Copy project files for installation
COPY --chown=armies:armies . /armies-src

# Install the armies package in editable mode
RUN pip install --no-cache-dir -e /armies-src

# Switch to non-root user
USER armies

# Profile data lives in a volume mount — never inside the image
VOLUME ["/home/armies/.armies"]

ENV HOME=/home/armies

ENTRYPOINT ["armies"]
CMD ["--help"]
