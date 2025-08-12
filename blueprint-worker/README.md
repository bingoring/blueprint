# Blueprint Worker Server

## 🏗️ 아키텍처 개요

Blueprint Worker Server는 메인 애플리케이션 서버에서 분리된 별도의 워커 서버로, 다음과 같은 비동기 작업들을 처리합니다:

## 📋 주요 기능

### 1. 📧 이메일 서비스 (`email_queue`)
- **이메일 인증 코드 발송**: 회원가입, 로그인 시 이메일 인증
- **직장 이메일 인증**: 레벨 2 신원 증명용 직장 이메일 인증
- **알림 이메일**: 프로젝트 업데이트, 거래 완료 등

### 2. 📱 SMS 서비스 (`sms_queue`)
- **휴대폰 본인인증**: PASS/SKT/KT/LG U+ 연동
- **SMS 인증 코드**: 6자리 인증번호 발송
- **알림 SMS**: 중요 거래 알림 등

### 3. 📁 파일 처리 서비스 (`file_processing_queue`)
- **서류 업로드**: 전문 자격증, 학위 증명서 등 스캔 파일
- **이미지 최적화**: 프로필 사진, 프로젝트 이미지 리사이징
- **파일 저장**: AWS S3/CloudFlare R2 등 클라우드 스토리지

### 4. 🔍 신원 증명 서비스 (`verification_queue`)
- **소셜 미디어 연동**: LinkedIn, GitHub, Twitter API 연동
- **외부 API 호출**: 회사 도메인 검증, 프로필 정보 확인
- **서류 검토**: AI를 통한 1차 서류 검토 (OCR + 유효성 검사)

## 🚀 워커 실행 방식

### Redis Streams 기반 큐 시스템
```bash
# 메인 서버에서 작업 전송
XADD email_queue * job_data '{"type":"send_email","to":"user@example.com",...}'

# 워커 서버에서 작업 소비
XREADGROUP GROUP email_workers worker_1 COUNT 1 BLOCK 5000 STREAMS email_queue >
```

### Consumer Group 패턴
- **고가용성**: 여러 워커 인스턴스가 동일한 큐를 처리
- **자동 장애복구**: 워커가 다운되면 다른 워커가 작업 인계
- **재시도 메커니즘**: 실패한 작업은 자동으로 재시도

## 📊 모니터링 & 로깅

### 큐 상태 모니터링
- **대기 중인 작업 수**: `XLEN queue_name`
- **처리 중인 작업**: `XPENDING queue_name group_name`
- **워커 상태**: Health check endpoint

### 로그 구조
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "worker": "email_worker_1",
  "job_id": "1641123000-0",
  "job_type": "send_email",
  "duration_ms": 245,
  "status": "success"
}
```

## 🔧 배포 방식

### Docker 컨테이너
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o worker ./cmd/worker

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/worker /usr/local/bin/
CMD ["worker"]
```

### 환경별 스케일링
- **개발환경**: 단일 워커 인스턴스
- **스테이징**: 2-3개 워커 인스턴스
- **프로덕션**: Auto-scaling (부하에 따라 자동 확장)

## 🛡️ 보안 고려사항

### 민감 정보 처리
- **이메일 내용**: 암호화하여 Redis에 저장
- **파일 업로드**: 바이러스 검사 후 저장
- **API 키 관리**: Vault/K8s Secret으로 관리

### 요청 제한
- **이메일**: 사용자당 시간당 10회
- **SMS**: 사용자당 일간 5회
- **파일 업로드**: 사용자당 일간 10MB

## 📈 성능 최적화

### 배치 처리
- **이메일**: 템플릿별로 배치 발송
- **이미지 처리**: 여러 이미지 동시 처리
- **API 호출**: 벌크 요청으로 최적화

### 캐싱 전략
- **템플릿 캐싱**: 이메일/SMS 템플릿 메모리 캐시
- **API 응답**: 외부 API 응답 Redis 캐시
- **파일 메타데이터**: 처리된 파일 정보 캐시

## 🚦 에러 처리 전략

### 재시도 정책
- **일시적 오류**: 지수 백오프로 3회 재시도
- **영구적 오류**: Dead Letter Queue로 이동
- **Critical 오류**: 즉시 알림 + 로그

### Circuit Breaker
- **외부 서비스 장애** 시 자동으로 요청 차단
- **복구 감지** 시 점진적으로 요청 재개
