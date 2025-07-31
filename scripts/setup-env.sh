#!/bin/bash

# 🌍 LifePathDAO Environment Setup Script
# This script helps you set up your .env file interactively

set -e

echo "🚀 LifePathDAO Environment Setup"
echo "=================================="
echo ""

# .env 파일 경로
ENV_FILE=".env"
ENV_EXAMPLE_FILE=".env.example"

# .env 파일 존재 확인
if [ -f "$ENV_FILE" ]; then
    echo "⚠️  .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "✅ Keeping existing .env file"
        exit 0
    fi
fi

# .env.example에서 복사
if [ -f "$ENV_EXAMPLE_FILE" ]; then
    echo "📁 Copying from .env.example..."
    cp "$ENV_EXAMPLE_FILE" "$ENV_FILE"
else
    echo "📝 Creating new .env file..."
    cat > "$ENV_FILE" << 'EOF'
# 🌍 LifePathDAO Environment Configuration
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=blueprint
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Server Configuration
PORT=8080
GIN_MODE=debug

# Frontend URL
FRONTEND_URL=http://localhost:3000

# OpenAI Configuration
OPENAI_API_KEY=your-openai-api-key
OPENAI_MODEL=gpt-4o-mini
EOF
fi

echo "✅ .env file created successfully!"
echo ""

# 선택적 설정 가이드
echo "🔧 Configuration Guide:"
echo "========================"
echo ""
echo "1. 🗄️  Database (PostgreSQL):"
echo "   Current: postgres/password@localhost:5432/blueprint"
echo "   ⚡ For local development, these defaults should work with docker-compose"
echo ""

echo "2. 🔑 JWT Secret:"
echo "   Current: Default secret (CHANGE IN PRODUCTION!)"
echo "   💡 Generate a secure secret: openssl rand -base64 32"
echo ""

echo "3. 🔗 Google OAuth (Optional):"
echo "   Current: Placeholder values"
echo "   📋 Setup guide: docs/google-oauth-setup.md"
echo ""

echo "4. 🤖 OpenAI API (Optional for AI features):"
echo "   Current: Placeholder value"
echo "   🔗 Get your key from: https://platform.openai.com/api-keys"
echo ""

# 대화형 설정 옵션
read -p "🛠️  Do you want to configure values interactively? (y/N): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "🔧 Interactive Configuration:"
    echo "============================="

    # JWT Secret
    read -p "Enter JWT secret (or press Enter to generate): " jwt_secret
    if [ -z "$jwt_secret" ]; then
        if command -v openssl >/dev/null 2>&1; then
            jwt_secret=$(openssl rand -base64 32)
            echo "🎲 Generated JWT secret: $jwt_secret"
        else
            jwt_secret="your-super-secret-jwt-key-$(date +%s)"
            echo "⚠️  OpenSSL not found. Using timestamp-based secret."
        fi
    fi
    sed -i.bak "s/JWT_SECRET=.*/JWT_SECRET=$jwt_secret/" "$ENV_FILE" && rm "$ENV_FILE.bak"

    # Google OAuth
    echo ""
    read -p "Enter Google Client ID (optional): " google_client_id
    if [ ! -z "$google_client_id" ]; then
        sed -i.bak "s/GOOGLE_CLIENT_ID=.*/GOOGLE_CLIENT_ID=$google_client_id/" "$ENV_FILE" && rm "$ENV_FILE.bak"

        read -p "Enter Google Client Secret: " google_client_secret
        if [ ! -z "$google_client_secret" ]; then
            sed -i.bak "s/GOOGLE_CLIENT_SECRET=.*/GOOGLE_CLIENT_SECRET=$google_client_secret/" "$ENV_FILE" && rm "$ENV_FILE.bak"
        fi
    fi

    # OpenAI API Key
    echo ""
    read -p "Enter OpenAI API Key (optional): " openai_key
    if [ ! -z "$openai_key" ]; then
        sed -i.bak "s/OPENAI_API_KEY=.*/OPENAI_API_KEY=$openai_key/" "$ENV_FILE" && rm "$ENV_FILE.bak"
    fi

    echo ""
    echo "✅ Configuration completed!"
fi

echo ""
echo "🎯 Next Steps:"
echo "=============="
echo "1. Review and edit .env file if needed: nano .env"
echo "2. Start development environment: make run-dev"
echo "3. Run backend in another terminal: make run-backend"
echo "4. Run frontend in another terminal: make run-frontend"
echo ""
echo "🔗 Useful links:"
echo "- Google OAuth setup: docs/google-oauth-setup.md"
echo "- OpenAI Platform: https://platform.openai.com/"
echo ""
echo "🚀 Happy coding!"
