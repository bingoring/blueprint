#!/bin/bash

# 로컬 개발 환경 변수 설정
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=blueprint_db
export DB_SSLMODE=disable

export JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

export GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID:-your-google-client-id}
export GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET:-your-google-client-secret}
export GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

export PORT=8080
export GIN_MODE=debug
export FRONTEND_URL=http://localhost:3000

# OpenAI API 설정 🤖
export OPENAI_API_KEY=${OPENAI_API_KEY:-your-openai-api-key}
export OPENAI_MODEL=gpt-4o-mini

echo "✅ 환경 변수가 설정되었습니다!"
echo "🤖 OpenAI 모델: $OPENAI_MODEL"
echo "🔗 API 키 설정됨: ${OPENAI_API_KEY:0:10}..."
