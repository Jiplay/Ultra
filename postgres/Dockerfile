FROM postgres:15-alpine

# Set environment variables
ENV POSTGRES_DB=ultra_food_db
ENV POSTGRES_USER=ultra_user
ENV POSTGRES_PASSWORD=ultra_password

# Copy initialization scripts
COPY init/ /docker-entrypoint-initdb.d/

# Create a directory for custom configurations
RUN mkdir -p /usr/share/postgresql/postgresql.conf.sample.d

# Expose PostgreSQL port
EXPOSE 5432

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pg_isready -U $POSTGRES_USER -d $POSTGRES_DB || exit 1

# The official postgres image already has the correct entrypoint
# No need to override CMD or ENTRYPOINT