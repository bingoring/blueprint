# 테스트 계정 생성 스크립트

Blueprint 서비스 테스팅을 위한 테스트 계정을 생성하는 스크립트입니다.

## 기본 사용법

### 1. 기본 테스트 계정 생성 (100명)
```bash
make create-test-accounts
```

### 2. 대량 부하 테스트용 계정 생성 (1000명)
```bash
make create-load-test-accounts
```

### 3. 기존 계정 삭제 후 재생성
```bash
make recreate-test-accounts
```

### 4. 실제 PostgreSQL 데이터베이스에 테스트 계정 생성
```bash
# 기본 테스트 계정 생성 (PostgreSQL)
make create-test-accounts-postgres

# 대량 부하 테스트용 계정 생성 (PostgreSQL)
make create-load-test-accounts-postgres

# 기존 계정 삭제 후 재생성 (PostgreSQL)
make recreate-test-accounts-postgres
```

## 고급 사용법

### 환경변수를 통한 커스터마이징
```bash
# 사용자 수 설정
NUM_USERS=500 make create-test-accounts

# USDC 잔액 설정 ($10,000 = 1000000 cents)
USDC_BALANCE=1000000 make create-test-accounts

# PostgreSQL 사용 (환경변수 자동 로드)
DB_TYPE=postgres make create-test-accounts

# PostgreSQL 사용 (수동 연결 정보 지정)
DB_TYPE=postgres DATABASE_URL="postgres://user:pass@localhost/dbname" make create-test-accounts

# 기존 계정 삭제
CLEAN_EXISTING=true make create-test-accounts
```

### 직접 스크립트 실행
```bash
cd blueprint-be

# 기본 실행
go run scripts/create_test_accounts.go

# 환경변수 설정 후 실행
NUM_USERS=200 USDC_BALANCE=5000000 go run scripts/create_test_accounts.go
```

## 환경변수 설명

| 변수명 | 기본값 | 설명 |
|--------|--------|------|
| `DB_TYPE` | `sqlite` | 데이터베이스 타입 (`sqlite`, `postgres`) |
| `DATABASE_URL` | `test.db` | 데이터베이스 연결 URL |
| `NUM_USERS` | `100` | 생성할 사용자 수 |
| `USDC_BALANCE` | `100000000` | 기본 USDC 잔액 (센트 단위, $1,000,000) |
| `CLEAN_EXISTING` | `false` | 기존 테스트 계정 삭제 여부 |

## 생성되는 계정 정보

### 사용자 계정
- **Username**: `testuser_1`, `testuser_2`, ..., `testuser_N`
- **Email**: `testuser1@example.com`, `testuser2@example.com`, ...
- **USDC 잔액**: 환경변수로 설정 (기본 $1,000,000)

### 테스트 프로젝트
- **Title**: "Test Project"
- **Status**: "active"
- **Owner**: testuser_1

### 테스트 마일스톤
- **Title**: "Test Milestone"
- **Status**: "funding"
- **Project**: Test Project

## 사용 예시

### API 테스트
```bash
# 주문 생성 테스트
curl -X POST http://localhost:8080/api/v1/trading/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "milestone_id": 1,
    "option_id": "success",
    "side": "buy",
    "quantity": 100,
    "price": 0.75
  }'
```

### 부하 테스트
```bash
# 테스트 계정 생성 후 부하 테스트 실행
make create-load-test-accounts
make test-load
```

### 통합 테스트
```bash
# 기본 계정으로 통합 테스트
make create-test-accounts
make test-integration
```

## 트러블슈팅

### 1. 데이터베이스 연결 오류
```bash
# PostgreSQL 연결 확인
psql -d $DATABASE_URL -c "SELECT 1;"

# SQLite 파일 권한 확인
ls -la test.db
```

### 2. 테이블 생성 오류
```bash
# 마이그레이션 수동 실행
go run cmd/migrate/main.go
```

### 3. 중복 계정 오류
```bash
# 기존 계정 삭제 후 재생성
CLEAN_EXISTING=true make create-test-accounts
```

### 4. 메모리 부족
```bash
# 사용자 수 줄이기
NUM_USERS=50 make create-test-accounts
```

## 개발자 팁

### 1. 빠른 개발용 설정
```bash
# 소수의 계정으로 빠른 테스트
NUM_USERS=10 USDC_BALANCE=10000000 make create-test-accounts
```

### 2. 성능 테스트용 설정
```bash
# 대량 계정으로 성능 테스트
NUM_USERS=10000 USDC_BALANCE=1000000000 make create-load-test-accounts
```

### 3. 프로덕션 유사 환경
```bash
# PostgreSQL + 실제 규모 데이터
make create-load-test-accounts-postgres

# 또는 직접 환경변수 설정
DB_TYPE=postgres NUM_USERS=1000 make create-test-accounts
```

## 스크립트 구조

```
scripts/
├── create_test_accounts.go    # 메인 스크립트
└── README.md                  # 이 파일
```

### 주요 기능
1. **데이터베이스 연결**: SQLite/PostgreSQL 지원
2. **테이블 마이그레이션**: 자동 스키마 생성
3. **사용자 생성**: 배치 처리로 효율적 생성
4. **지갑 초기화**: 각 사용자별 USDC 잔액 설정
5. **테스트 데이터**: 프로젝트/마일스톤 자동 생성
6. **진행률 표시**: 대량 생성 시 진행률 모니터링

이 스크립트를 통해 Blueprint 서비스의 다양한 테스트 시나리오를 효과적으로 수행할 수 있습니다.