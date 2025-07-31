#!/bin/bash

# Blueprint 환경변수 설정 스크립트

echo "🔧 Blueprint 환경변수 설정 중..."

# 데이터베이스 설정
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=blueprint_db
export DB_SSLMODE=disable

# JWT 설정
export JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Google OAuth 설정
# 프로젝트 ID: blueprint-467515 (Service Account에서 확인)
export GOOGLE_PROJECT_ID=blueprint-467515

# ✅ OAuth 2.0 Client ID 설정 완료!
export GOOGLE_CLIENT_ID="${GOOGLE_CLIENT_ID:-your-google-client-id}"
export GOOGLE_CLIENT_SECRET="${GOOGLE_CLIENT_SECRET:-your-google-client-secret}"
export GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# 서버 설정
export PORT=8080
export GIN_MODE=debug
export FRONTEND_URL=http://localhost:3000

# Google OAuth 파일 경로
export GOOGLE_SERVICE_ACCOUNT_FILE=./blueprint-467515-134e003cd7f4.json
export GOOGLE_OAUTH_CLIENT_FILE=./oauth-client-secret.json

echo "✅ 환경변수 설정 완료!"
echo ""
echo "📝 설정된 환경변수:"
echo "   - DB_NAME: $DB_NAME"
echo "   - GOOGLE_PROJECT_ID: $GOOGLE_PROJECT_ID"
echo "   - GOOGLE_CLIENT_ID: $GOOGLE_CLIENT_ID"
echo "   - PORT: $PORT"
echo ""
echo "✅ Google OAuth 2.0 설정 완료!"
echo "   - 클라이언트 ID: $GOOGLE_CLIENT_ID"
echo "   - 리디렉션 URI: $GOOGLE_REDIRECT_URL"
echo "   - JavaScript Origins: http://localhost:3000"
