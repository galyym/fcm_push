#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== FCM Push Service - Development Setup ===${NC}\n"

if [ ! -f .env ]; then
    echo -e "${YELLOW}Warning: .env file not found${NC}"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo -e "${RED}Please edit .env file with your Firebase credentials${NC}"
    exit 1
fi

echo "Loading environment variables..."
export $(cat .env | grep -v '^#' | xargs)

if [ -z "$FCM_CREDENTIALS_PATH" ]; then
    echo -e "${RED}Error: FCM_CREDENTIALS_PATH is not set in .env${NC}"
    exit 1
fi

if [ -z "$FCM_PROJECT_ID" ]; then
    echo -e "${RED}Error: FCM_PROJECT_ID is not set in .env${NC}"
    exit 1
fi

if [ ! -f "$FCM_CREDENTIALS_PATH" ]; then
    echo -e "${RED}Error: Firebase credentials file not found at: $FCM_CREDENTIALS_PATH${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Environment variables loaded${NC}"
echo -e "${GREEN}✓ Firebase credentials found${NC}\n"

echo "Installing dependencies..."
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}\n"
echo -e "${GREEN}Starting FCM Push Service...${NC}"
echo -e "Server will be available at: ${YELLOW}http://localhost:${SERVER_PORT:-8080}${NC}\n"

go run cmd/server/main.go
