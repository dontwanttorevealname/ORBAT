#!/bin/bash

if [ -f .env.test ]; then
    source .env.test
fi

cd SQL/Migrations
echo "Resetting database..."
# Remove the goose version table to force a clean start
goose turso "$DATABASE_URL" down-to 0
goose turso "$DATABASE_URL" reset

echo "Running migrations..."
# Force all migrations to run from the beginning
goose turso "$DATABASE_URL" up

cd ../Seeds
echo "Running seeds..."
goose turso "$DATABASE_URL" -no-versioning up

echo "Migrations and seeds completed successfully"