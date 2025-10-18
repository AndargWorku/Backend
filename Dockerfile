# Use Hasura's official image
FROM hasura/graphql-engine:v2.38.0

# Copy the Hasura metadata, migrations, and config if you have them
# (optional — only if your repo includes migrations/metadata)
COPY ./migrations /hasura-migrations
COPY ./metadata /hasura-metadata

# Expose Hasura's port
EXPOSE 8080

# The CMD defines how the service runs — Render uses this automatically
CMD ["graphql-engine", "serve"]
