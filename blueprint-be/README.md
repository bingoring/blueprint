# Blueprint Backend

Go로 구축된 Blueprint 플랫폼의 백엔드 API 서버입니다.

## 🛠️ 기술 스택

- **Go 1.21+**: 백엔드 언어
- **Gin**: HTTP 웹 프레임워크
- **GORM**: ORM (PostgreSQL)
- **JWT**: 인증/권한
- **Redis**: 캐싱 및 실시간 데이터
- **PostgreSQL**: 메인 데이터베이스
- **TimescaleDB**: 시계열 데이터 (거래 히스토리)
- **Docker**: 컨테이너화

## 🚀 시작하기

### 1. 의존성 설치
```bash
go mod download
```

### 2. 데이터베이스 시작
```bash
make dev-db
```

### 3. 서버 실행
```bash
go run cmd/server/main.go
```

또는 개발 환경 전체 시작:
```bash
make dev-start
```

## 📁 프로젝트 구조

```
cmd/
└── server/           # 메인 애플리케이션
internal/
├── config/           # 설정 관리
├── database/         # 데이터베이스 연결
├── handlers/         # HTTP 핸들러
├── middleware/       # 미들웨어
├── models/           # 데이터 모델
├── services/         # 비즈니스 로직
└── queue/            # 비동기 작업 큐
pkg/
├── utils/            # 유틸리티
└── validation/       # 검증 로직
```

## 🌐 주요 기능

### 인증 시스템
- JWT 기반 인증
- Google OAuth 2.0
- 리프레시 토큰

### 거래 시스템
- **고성능 매칭 엔진**: Price-Time Priority
- **실시간 업데이트**: Server-Sent Events (SSE)
- **Market Maker**: 자동 유동성 공급
- **P2P 베팅**: 폴리마켓 스타일

### 데이터 관리
- **GORM**: 자동 마이그레이션
- **Redis**: 실시간 캐싱
- **TimescaleDB**: 거래 히스토리

## 🔧 환경 설정

### 환경 변수 (.env)
```bash
# 데이터베이스
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=blueprint

# JWT
JWT_SECRET=your-secret-key

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
```

## 📊 API 엔드포인트

### 인증
- `POST /api/v1/auth/register` - 회원가입
- `POST /api/v1/auth/login` - 로그인
- `GET /api/v1/auth/google/login` - Google 로그인

### 프로젝트
- `GET /api/v1/projects` - 프로젝트 목록
- `POST /api/v1/projects` - 프로젝트 생성
- `GET /api/v1/projects/:id` - 프로젝트 조회

### 거래
- `POST /api/v1/orders` - 주문 생성
- `GET /api/v1/milestones/:id/orderbook/:option` - 호가창
- `GET /api/v1/milestones/:id/stream` - 실시간 SSE

## 🐳 Docker

### 개발 환경
```bash
make dev-db        # 데이터베이스만 시작
make dev-start     # 전체 개발 환경
```

### 운영 환경
```bash
make build         # 이미지 빌드
make up           # 전체 서비스 시작
```

## 💀 데이터 초기화

```bash
make nuke-db      # 데이터베이스만 초기화
make nuke-all     # 모든 데이터 초기화
make fresh-start  # 완전 재시작
```

## 🔍 로깅

- **Info 레벨**: 기본 애플리케이션 로그
- **Error 레벨**: 데이터베이스 쿼리 (실패만)
- **Debug 레벨**: 상세 디버그 정보

## 🚀 성능

- **매칭 엔진**: 10,000+ 주문/초 처리 가능
- **SSE**: 동시 연결 1,000+ 클라이언트
- **Redis 캐싱**: 밀리초 단위 응답
- **고루틴**: 비동기 처리로 높은 동시성
