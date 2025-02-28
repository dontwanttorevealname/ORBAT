#!/bin/bash

if [ -f .env.test ]; then
    source .env.test
fi

cd SQL/Migrations
echo "Resetting database..."
goose turso "$DATABASE_URL" reset

echo "Running migrations..."
goose turso "$DATABASE_URL" up

cd ../Seeds
echo "Running seeds..."
goose turso "$DATABASE_URL" -no-versioning up

echo "Migrations and seeds completed successfully"