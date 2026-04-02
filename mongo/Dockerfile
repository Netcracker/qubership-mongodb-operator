ARG MONGO_VERSION=8.0.20
FROM mongo:${MONGO_VERSION}

USER root

# Remove the existing mongodb user/group if needed (optional, careful with this)
RUN deluser mongodb || true && delgroup mongodb || true

# Create mongodb user/group with UID and GID 1001
RUN groupadd --gid 1001 mongodb && \
    useradd --uid 1001 --gid mongodb --home-dir /data/db --shell /bin/bash --system mongodb

# Ensure ownership of data directories
RUN mkdir -p /data/db /data/configdb && \
    chown -R mongodb:mongodb /data/db /data/configdb

# Copy original entrypoint script to modify user from 999 to 1001
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Set user back to root so gosu can switch to mongodb (uid 1001)
USER root

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["mongod"]