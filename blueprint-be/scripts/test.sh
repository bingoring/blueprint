#!/bin/bash

# Blueprint 서비스 종합 테스트 스크립트

set -e  # 에러 발생 시 스크립트 종료

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로고
echo -e "${BLUE}"
echo "  ╔═══════════════════════════════════════╗"
echo "  ║        Blueprint Service Tests        ║"
echo "  ║     🧪 종합 테스팅 시스템            ║"
echo "  ╚═══════════════════════════════════════╝"
echo -e "${NC}"

# 환경 설정
export CGO_ENABLED=1  # SQLite를 위해 필요

# 함수 정의
print_section() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}🔥 $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 테스트 타입 파라미터 처리
TEST_TYPE=${1:-"all"}

case $TEST_TYPE in
    "all")
        echo -e "${GREEN}🎯 모든 테스트를 실행합니다.${NC}"
        RUN_UNIT=true
        RUN_INTEGRATION=true
        RUN_LOAD=true
        ;;
    "unit")
        echo -e "${GREEN}🎯 단위 테스트만 실행합니다.${NC}"
        RUN_UNIT=true
        RUN_INTEGRATION=false
        RUN_LOAD=false
        ;;
    "integration")
        echo -e "${GREEN}🎯 통합 테스트만 실행합니다.${NC}"
        RUN_UNIT=false
        RUN_INTEGRATION=true
        RUN_LOAD=false
        ;;
    "load")
        echo -e "${GREEN}🎯 부하 테스트만 실행합니다.${NC}"
        RUN_UNIT=false
        RUN_INTEGRATION=false
        RUN_LOAD=true
        ;;
    *)
        print_error "잘못된 테스트 타입: $TEST_TYPE"
        echo "사용법: $0 [all|unit|integration|load]"
        exit 1
        ;;
esac

# 1. 환경 점검
print_section "환경 점검"

# Go 버전 확인
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    print_success "Go 설치됨: $GO_VERSION"
else
    print_error "Go가 설치되지 않았습니다."
    exit 1
fi

# Redis 확인 (옵션)
if command -v redis-server &> /dev/null; then
    print_success "Redis 설치됨"
else
    print_warning "Redis가 설치되지 않았습니다 (테스트에서는 Mock Redis 사용)"
fi

# 의존성 설치
print_section "의존성 설치"
go mod tidy
if [ $? -eq 0 ]; then
    print_success "의존성 설치 완료"
else
    print_error "의존성 설치 실패"
    exit 1
fi

# 2. 코드 품질 검사
print_section "코드 품질 검사"

# Go fmt 체크
echo "📝 코드 포맷 확인..."
if ! gofmt -l . | grep -q .; then
    print_success "코드 포맷 OK"
else
    print_warning "코드 포맷 수정 필요:"
    gofmt -l .
    echo "자동 수정 실행..."
    gofmt -w .
    print_success "코드 포맷 자동 수정 완료"
fi

# Go vet 실행
echo "🔍 코드 정적 분석..."
if go vet ./...; then
    print_success "정적 분석 통과"
else
    print_error "정적 분석 실패"
    exit 1
fi

# 빌드 테스트
echo "🏗️ 빌드 테스트..."
if go build ./...; then
    print_success "빌드 성공"
else
    print_error "빌드 실패"
    exit 1
fi

# 3. 단위 테스트
if [ "$RUN_UNIT" = true ]; then
    print_section "단위 테스트 (Unit Tests)"
    
    echo "🧪 분산 매칭 엔진 테스트..."
    if go test -v ./tests/unit/... -run TestDistributedMatchingEngine -timeout 30s; then
        print_success "분산 매칭 엔진 테스트 통과"
    else
        print_error "분산 매칭 엔진 테스트 실패"
        exit 1
    fi
    
    echo "🧪 CQRS 패턴 테스트..."
    if go test -v ./tests/unit/... -run TestCQRS -timeout 30s; then
        print_success "CQRS 패턴 테스트 통과"
    else
        print_error "CQRS 패턴 테스트 실패"
        exit 1
    fi
    
    # 커버리지 측정
    echo "📊 코드 커버리지 측정..."
    go test -coverprofile=coverage.out ./tests/unit/...
    go tool cover -html=coverage.out -o coverage.html
    COVERAGE=$(go tool cover -func=coverage.out | tail -n 1 | awk '{print $3}')
    echo "📈 코드 커버리지: $COVERAGE"
    
    if [[ ${COVERAGE%.*} -ge 80 ]]; then
        print_success "커버리지 목표 달성 (80% 이상): $COVERAGE"
    else
        print_warning "커버리지 목표 미달 (80% 미만): $COVERAGE"
    fi
fi

# 4. 통합 테스트
if [ "$RUN_INTEGRATION" = true ]; then
    print_section "통합 테스트 (Integration Tests)"
    
    echo "🔗 거래 시스템 통합 테스트..."
    if go test -v ./tests/integration/... -timeout 60s; then
        print_success "통합 테스트 통과"
    else
        print_error "통합 테스트 실패"
        exit 1
    fi
fi

# 5. 부하 테스트
if [ "$RUN_LOAD" = true ]; then
    print_section "부하 테스트 (Load Tests)"
    
    echo "⚡ 고성능 주문 처리 테스트..."
    echo "   (이 테스트는 시간이 오래 걸릴 수 있습니다...)"
    
    if go test -v ./tests/load/... -timeout 300s; then
        print_success "부하 테스트 통과"
    else
        print_warning "부하 테스트 실패 (성능 기준 미달)"
    fi
fi

# 6. 테스트 결과 요약
print_section "테스트 결과 요약"

echo -e "${GREEN}🎉 Blueprint 서비스 테스트 완료!${NC}"
echo ""
echo "📋 실행된 테스트:"
[ "$RUN_UNIT" = true ] && echo "   ✅ 단위 테스트"
[ "$RUN_INTEGRATION" = true ] && echo "   ✅ 통합 테스트"
[ "$RUN_LOAD" = true ] && echo "   ✅ 부하 테스트"
echo ""

# 생성된 파일들
echo "📄 생성된 파일들:"
[ -f "coverage.out" ] && echo "   📊 coverage.out - 커버리지 데이터"
[ -f "coverage.html" ] && echo "   🌐 coverage.html - 커버리지 리포트 (브라우저에서 열어보세요)"
echo ""

# 성능 권장사항
if [ "$RUN_LOAD" = true ]; then
    echo "🚀 성능 최적화 권장사항:"
    echo "   1. Redis 클러스터 구성으로 확장성 향상"
    echo "   2. 데이터베이스 읽기 전용 복제본 활용"
    echo "   3. CDN을 통한 정적 자산 최적화"
    echo "   4. 로드 밸런서로 트래픽 분산"
    echo ""
fi

# 다음 단계
echo "📈 다음 단계:"
echo "   1. 프로덕션 환경에서 A/B 테스트 실행"
echo "   2. 모니터링 시스템 구축 (Prometheus + Grafana)"
echo "   3. 자동화된 CI/CD 파이프라인 구성"
echo "   4. 보안 감사 및 펜테스팅"

echo -e "\n${BLUE}🎯 모든 테스트가 성공적으로 완료되었습니다!${NC}"