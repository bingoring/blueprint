# Blueprint 서비스 테스팅 가이드 🧪

Blueprint 분산 거래 시스템의 포괄적인 테스팅 전략과 실행 가이드입니다.

## 📋 테스트 개요

### 🎯 테스팅 목표
- **안정성**: 분산 환경에서의 시스템 안정성 보장
- **성능**: 고부하 상황에서의 성능 검증 
- **정확성**: 거래 로직의 정확성 및 데이터 일관성 확인
- **확장성**: 다중 서버 환경에서의 확장성 테스트
- **실제 시나리오**: 현실적인 사용 사례 검증

### 🏗️ 테스트 아키텍처
```
tests/
├── unit/                   # 단위 테스트
│   ├── distributed_matching_engine_test.go
│   └── cqrs_test.go
├── integration/            # 통합 테스트  
│   └── trading_integration_test.go
├── load/                   # 부하 테스트
│   └── matching_engine_load_test.go
├── e2e/                    # E2E 테스트
│   └── real_world_scenarios_test.go
└── fixtures/              # 테스트 데이터
```

## 🚀 빠른 시작

### 전체 테스트 실행
```bash
# 모든 테스트 실행 (권장)
./scripts/test.sh all

# 또는 개별 실행
go test -v ./tests/...
```

### 테스트 타입별 실행
```bash
# 단위 테스트만
./scripts/test.sh unit

# 통합 테스트만  
./scripts/test.sh integration

# 부하 테스트만
./scripts/test.sh load
```

### 코드 커버리지 확인
```bash
go test -coverprofile=coverage.out ./tests/unit/...
go tool cover -html=coverage.out -o coverage.html
```

## 📊 테스트 분류

### 1. 단위 테스트 (Unit Tests) 🧪

**목적**: 개별 컴포넌트의 로직 검증

**커버리지**:
- ✅ 분산 매칭 엔진 핵심 로직
- ✅ Redis 기반 이벤트 소싱
- ✅ 분산 락 메커니즘 (Redlock)
- ✅ CQRS 명령/조회 분리
- ✅ 가격 오라클 업데이트
- ✅ 주문장 관리

**실행 시간**: ~30초

**주요 테스트 케이스**:
```go
// 주문 매칭 테스트
func TestOrderMatching()

// 동시 주문 처리 테스트  
func TestConcurrentOrders()

// 분산 락 테스트
func TestDistributedLocking()

// 이벤트 소싱 테스트
func TestEventSourcing()
```

### 2. 통합 테스트 (Integration Tests) 🔗

**목적**: 시스템 컴포넌트 간 상호작용 검증

**커버리지**:
- ✅ API 엔드포인트 통합
- ✅ 데이터베이스 연동
- ✅ Redis 캐시 연동  
- ✅ SSE 실시간 스트리밍
- ✅ 거래 시스템 전체 플로우

**실행 시간**: ~60초

**주요 테스트 케이스**:
```go
// 주문 매칭 통합 테스트
func TestOrderMatching()

// 동시 거래 테스트
func TestConcurrentTrading()

// SSE 스트리밍 테스트
func TestSSEStreaming()

// 데이터 일관성 테스트
func TestDataConsistency()
```

### 3. 부하 테스트 (Load Tests) ⚡

**목적**: 고부하 상황에서의 성능 및 안정성 검증

**성능 목표**:
- 📈 **처리율**: 1000+ orders/sec
- 📊 **응답시간**: 평균 < 200ms
- 🎯 **성공률**: > 95%
- 💾 **메모리**: 안정적 사용량

**실행 시간**: ~5분

**주요 테스트 케이스**:
```go
// 대량 주문 처리 (1000+ orders)
func TestHighVolumeOrderProcessing()

// 동시 접근 테스트 (100 concurrent users)
func TestConcurrentMarketAccess()

// 메모리 사용량 테스트
func TestMemoryUsage()

// Redis 연결 풀 테스트
func TestRedisConnectionPool()
```

### 4. E2E 테스트 (End-to-End Tests) 🌐

**목적**: 실제 사용 시나리오 기반 검증

**실행 시간**: ~2분

**실제 시나리오**:

#### 📈 스타트업 마일스톤 거래
```
1️⃣ 초기 시장 설정 (50/50 확률)
2️⃣ 내부 정보 보유자의 대량 베팅  
3️⃣ 시장 반응 및 가격 변동
4️⃣ 추가 투자자 유입
5️⃣ 실제 이벤트 발생 시뮬레이션
```

#### 📱 인플루언서 프로젝트 거래
```
1️⃣ 팬덤의 낙관적 베팅
2️⃣ 데이터 분석가의 냉정한 분석
3️⃣ 중립적 투자자들의 차익거래
4️⃣ 시장 효율성 검증
```

#### 🛡️ 시장 조작 방어
```
1️⃣ 정상적인 시장 형성
2️⃣ Pump & Dump 조작 시도
3️⃣ 시장의 자연스러운 방어
4️⃣ 조작 효과 무력화 확인
```

## 📋 테스트 체크리스트

### ✅ 기능 테스트
- [ ] 주문 생성/취소/매칭
- [ ] 실시간 가격 업데이트
- [ ] 사용자 잔액 관리
- [ ] 거래 내역 기록
- [ ] SSE 스트리밍
- [ ] API 응답 정확성

### ✅ 성능 테스트  
- [ ] 초당 1000+ 주문 처리
- [ ] 평균 응답시간 < 200ms
- [ ] 동시 접속자 100+ 지원
- [ ] 메모리 사용량 안정성
- [ ] 데이터베이스 연결 풀

### ✅ 안정성 테스트
- [ ] 분산 락 정확성
- [ ] 이벤트 소싱 일관성
- [ ] 다중 서버 동기화
- [ ] 오류 상황 복구
- [ ] 데이터 무결성

### ✅ 보안 테스트
- [ ] 사용자 인증/인가
- [ ] 잔액 검증
- [ ] SQL 인젝션 방어
- [ ] XSS 방어
- [ ] 시장 조작 방어

## 🔧 테스트 환경 설정

### 필수 의존성
```bash
# 테스트 라이브러리 설치
go get -t github.com/stretchr/testify/suite
go get -t github.com/stretchr/testify/assert  
go get -t github.com/alicebob/miniredis/v2
go get -t github.com/DATA-DOG/go-sqlmock
go get -t gorm.io/driver/sqlite
```

### 환경 변수
```bash
export CGO_ENABLED=1  # SQLite를 위해 필요
export GO_ENV=test    # 테스트 환경 설정
```

### Mock 서비스
- **Redis**: miniredis (In-memory Redis mock)
- **Database**: SQLite (In-memory)  
- **SSE**: Mock SSE 서비스

## 📊 성능 벤치마크

### 🎯 목표 성능 지표
| 지표 | 목표값 | 단위 |
|------|--------|------|
| 주문 처리율 | 1,000+ | orders/sec |
| 평균 응답시간 | < 200 | milliseconds |
| 동시 사용자 | 100+ | concurrent users |
| 성공률 | > 95% | percentage |
| CPU 사용률 | < 70% | percentage |
| 메모리 사용량 | < 500MB | per instance |

### 📈 실제 측정 결과 (예시)
```
📊 부하 테스트 결과:
   - 총 주문 수: 10,000
   - 성공: 9,850, 실패: 150
   - 소요 시간: 8.5s
   - 초당 주문 처리율: 1,176.47 orders/sec
   - 평균 응답 시간: 85.23 ms
   - 성공률: 98.5%
```

## 🚨 문제 해결

### 자주 발생하는 이슈

#### 1. Redis 연결 실패
```bash
# 해결책: miniredis 포트 충돌 확인
netstat -an | grep :6379
```

#### 2. SQLite 드라이버 오류  
```bash
# 해결책: CGO 활성화
export CGO_ENABLED=1
go test -v ./tests/...
```

#### 3. 메모리 부족
```bash
# 해결책: 테스트 데이터 크기 조절
go test -v ./tests/load/... -short
```

#### 4. 타임아웃 오류
```bash
# 해결책: 타임아웃 증가  
go test -v ./tests/... -timeout 300s
```

## 📚 참고 자료

### 🔗 관련 문서
- [분산 매칭 엔진 아키텍처](../docs/distributed-matching-engine.md)
- [CQRS 패턴 구현](../docs/cqrs-implementation.md)  
- [Redis 이벤트 소싱](../docs/event-sourcing.md)
- [성능 최적화 가이드](../docs/performance-optimization.md)

### 🛠️ 도구 및 라이브러리
- [testify](https://github.com/stretchr/testify) - Go 테스트 프레임워크
- [miniredis](https://github.com/alicebob/miniredis) - Redis Mock
- [sqlmock](https://github.com/DATA-DOG/go-sqlmock) - SQL Mock
- [Gin Test](https://gin-gonic.com/docs/testing/) - HTTP 테스트

## 🎯 CI/CD 통합

### GitHub Actions 예시
```yaml
name: Blueprint Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: Run Tests
        run: |
          export CGO_ENABLED=1
          ./scripts/test.sh all
      
      - name: Upload Coverage
        run: |
          bash <(curl -s https://codecov.io/bash)
```

## 🏆 베스트 프랙티스

### ✅ DO
- 테스트는 독립적이고 재현 가능하게 작성
- Mock을 활용하여 외부 의존성 제거
- 실제 사용 시나리오 기반 E2E 테스트
- 성능 기준점을 명확히 설정
- 실패 시나리오도 반드시 테스트

### ❌ DON'T  
- 실제 프로덕션 데이터 사용 금지
- 테스트 간 상태 공유 금지
- 하드코딩된 대기 시간 사용 금지
- 네트워크 의존적 테스트 작성 금지
- 테스트 코드에 비즈니스 로직 포함 금지

---

**🔥 이 테스팅 시스템으로 Blueprint 서비스의 안정성과 성능을 보장하세요!**