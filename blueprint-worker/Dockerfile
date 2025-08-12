# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 의존성 파일 복사
COPY go.mod go.sum ./
COPY ../blueprint-module/go.mod ../blueprint-module/go.sum ../blueprint-module/

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .
COPY ../blueprint-module ../blueprint-module

# 바이너리 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker

# Runtime stage
FROM alpine:latest

# 필요한 패키지 설치
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/worker .

# 업로드 디렉토리 생성
RUN mkdir -p ./uploads

# 포트 노출 (모니터링용, 실제로는 worker는 포트 사용 안함)
EXPOSE 8081

# 실행
CMD ["./worker"]
