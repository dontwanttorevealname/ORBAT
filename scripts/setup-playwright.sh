#!/bin/bash

# Install Playwright CLI
go install github.com/playwright-community/playwright-go/cmd/playwright@latest

# Install browser binaries
playwright install

# Make script executable
chmod +x scripts/setup-playwright.sh 