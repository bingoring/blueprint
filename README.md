# Blueprint

개인의 장기 목표 달성을 위해 전문가 집단이 데이터 기반으로 최적 경로를 제안하고, 성과에 따라 보상받는 분산형 라이프 코칭 플랫폼

## 📁 프로젝트 구조

```
blueprint/
├── blueprint-be/     # 백엔드 (Go)
├── blueprint-fe/     # 프론트엔드 (React + TypeScript)
├── module/          # 공통 모듈 (MSA 대비)
├── docker-compose.yml
└── README.md
```

## 🚀 기능

### Milestone 1 (Current)
- [x] 사용자 등록/로그인 시스템
- [x] Google OAuth 2.0 인증
- [x] JWT 기반 인증
- [x] 사용자 프로필 관리
- [x] **Docker 기반 개발 환경**
- [ ] 목표 설정 시스템
- [ ] 기본 경로 제안 기능

### Milestone 2 (Planned)
- [ ] 예측 마켓 시스템
- [ ] 토큰 이코노미
- [ ] 스마트 컨트랙트 연동

## 🛠 기술 스택

- **Backend**: Go, Gin Framework
- **Database**: PostgreSQL, GORM
- **Cache**: Redis
- **Authentication**: JWT, Google OAuth 2.0
- **Infrastructure**: Docker, Docker Compose
- **Future**: Ethereum/Polygon, Solidity

## 📋 사전 요구사항

- Docker 20.10+
- Docker Compose 2.0+
- (선택사항) Go 1.24+ (로컬 개발용)

## 🐳 Docker로 빠른 시작 (권장)

### 1. 저장소 클론
```bash
git clone <repository-url>
cd blueprint
```

### 2. 한 번에 모든 설정 및 실행
```bash
make install
```

이 명령어는 다음을 자동으로 수행합니다:
- Docker 이미지 빌드
- PostgreSQL, Redis, 애플리케이션 서비스 시작
- 데이터베이스 마이그레이션 자동 실행

### 3. 서비스 확인
```bash
# 서비스 상태 확인
make status

# 로그 확인
make logs

# 애플리케이션만 로그 확인
make logs-app
```

### 4. 서비스 접속
- **API 서버**: http://localhost:8080
- **헬스 체크**: http://localhost:8080/health
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## 🔧 Docker 명령어 모음

```bash
# 도움말 (모든 명령어 확인)
make help

# 개발 모드 (코드 변경 시 자동 rebuild)
make dev

# 서비스 시작/중지
make up          # 백그라운드에서 시작
make up-logs     # 로그와 함께 시작
make down        # 모든 서비스 중지

# 로그 확인
make logs        # 모든 서비스 로그
make logs-app    # 애플리케이션 로그만
make logs-db     # 데이터베이스 로그만

# 컨테이너 접속
make shell-app   # 애플리케이션 컨테이너 셸
make shell-db    # 데이터베이스 접속

# 정리
make clean       # 컨테이너, 볼륨 제거
make clean-all   # 모든 Docker 리소스 제거

# 데이터베이스 백업
make backup-db
```

## 🚀 로컬 개발 (Docker 없이)

### 1. 환경변수 설정 (선택사항)

**Google OAuth 설정**:
```bash
# 환경변수 설정 스크립트 실행
source scripts/setup-env.sh

# 또는 Makefile 사용
make setup-env
```

📋 **상세 설정 가이드**: [docs/google-oauth-setup.md](docs/google-oauth-setup.md)

**프로젝트 정보**:
- 프로젝트 ID: `blueprint-467515` ✅
- Service Account: 있음 (서버 간 통신용)
- OAuth Client ID: ✅ **설정 완료!**
- 클라이언트 ID: `475922118539-g8plhmjifnenttr36956q7a437ols7eq.apps.googleusercontent.com`

### 2. 의존성 설치
```bash
go mod tidy
```

### 2. PostgreSQL 설치 및 설정
```bash
# macOS
brew install postgresql
brew services start postgresql
createdb blueprint_db

# 또는 Docker로 PostgreSQL만 실행
docker run --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:16-alpine
```

### 3. 빠른 개발 환경 시작

**데이터베이스만 시작** (권장):
```bash
# PostgreSQL, Redis만 Docker로 시작
make dev-db

# 백엔드 로컬 실행
make run-backend

# 프론트엔드 로컬 실행 (다른 터미널)
make run-frontend
```

**환경변수 자동 설정**:
```bash
# 환경변수 설정 후 백엔드 시작
make run-backend-with-env
```

**수동 환경 변수 설정**:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=blueprint_db
export JWT_SECRET=your-super-secret-jwt-key
export GOOGLE_CLIENT_ID=your-google-client-id
export GOOGLE_CLIENT_SECRET=your-google-client-secret

# 서버 실행
go run cmd/server/main.go
```

### 4. 접속 주소
- **프론트엔드**: http://localhost:3000
- **백엔드 API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## 📡 API 엔드포인트

### 인증 (Authentication)
- `POST /api/v1/auth/register` - 회원가입
- `POST /api/v1/auth/login` - 로그인
- `GET /api/v1/auth/google/login` - Google OAuth 시작 ✅
- `GET /api/v1/auth/google/callback` - Google OAuth 콜백 ✅

### 사용자 (User)
- `GET /api/v1/me` - 현재 사용자 정보 조회 (인증 필요)

### 헬스 체크
- `GET /health` - 서버 상태 확인

## 🧪 API 테스트

### 회원가입
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123"
  }'
```

### 로그인
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 헬스 체크
```bash
curl http://localhost:8080/health
```

## 🏗 프로젝트 구조

```
blueprint/
├── cmd/
│   └── server/           # 메인 애플리케이션
├── internal/
│   ├── config/          # 설정 관리
│   ├── database/        # 데이터베이스 연결
│   ├── handlers/        # HTTP 핸들러
│   ├── middleware/      # 미들웨어
│   ├── models/          # 데이터 모델
│   └── services/        # 비즈니스 로직
├── pkg/
│   ├── utils/           # 유틸리티 함수
│   └── validation/      # 검증 로직
├── Dockerfile           # 애플리케이션 Docker 이미지
├── docker-compose.yml   # 전체 스택 정의
├── Makefile            # Docker 명령어 모음
├── init.sql            # 데이터베이스 초기화
└── docs/               # 문서
```

## 🔧 개발 팁

### 코드 변경 시 자동 재시작
```bash
# 개발 모드로 실행 (코드 변경 시 자동 rebuild)
make dev
```

### 데이터베이스 직접 접속
```bash
# 데이터베이스 컨테이너에 접속
make shell-db

# 또는 외부에서 접속
psql -h localhost -p 5432 -U postgres -d blueprint_db
```

### 로그 실시간 확인
```bash
# 모든 서비스 로그
make logs

# 특정 서비스만
make logs-app
make logs-db
```

## 🚀 개발 환경 설정

### 1. 데이터베이스 시작
```bash
docker-compose up -d postgres redis
```

### 2. 백엔드 실행
```bash
cd blueprint-be
go run cmd/server/main.go
```

### 3. 프론트엔드 실행
```bash
cd blueprint-fe
npm install
npm run dev
```

### 4. 전체 개발 환경 (백엔드만)
```bash
cd blueprint-be
make dev-start  # 데이터베이스 + 백엔드 시작
```

## 🔮 다음 단계

1. **목표 관리 시스템 구현**
   - 목표 생성/수정/삭제 API
   - 목표 카테고리별 분류
   - 우선순위 및 기한 관리

2. **경로 제안 시스템**
   - 전문가의 경로 제안 기능
   - 경로별 상세 단계 관리
   - 소요 시간/비용 예측

3. **예측 마켓 구현**
   - 전문가 베팅 시스템
   - 동적 확률 계산
   - 보상 분배 메커니즘

## 🤝 기여

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 라이센스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 📞 연락처

프로젝트 관련 문의사항이 있으시면 이슈를 생성해주세요.
