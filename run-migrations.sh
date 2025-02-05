#!/bin/bash

if [ -f .env ]; then
    source .env
fi

cd SQL/Migrations
goose turso "$DATABASE_URL" up

