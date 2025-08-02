# 🐛 Blueprint Backend Debugging Guide

이 문서는 VS Code에서 Blueprint 백엔드를 디버깅하는 방법을 안내합니다.

## 📋 사전 준비

### 1. 필수 확장 프로그램 설치
VS Code를 열면 다음 확장 프로그램 설치를 추천합니다:
- **Go** (golang.go) - Go 언어 지원
- **PostgreSQL** (ms-ossdata.vscode-postgresql) - 데이터베이스 관리
- **Thunder Client** (rangav.vscode-thunder-client) - API 테스트

### 2. 환경 설정
```bash
# .env 파일 생성 (루트 디렉토리)
cp .env.example .env

# 환경변수 설정
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=blueprint
SERVER_PORT=8080
GIN_MODE=debug
```

## 🚀 디버깅 실행 방법

### 방법 1: VS Code Debug Panel 사용

1. **F5** 키를 누르거나 **Run and Debug** 패널 열기
2. 디버그 설정 선택:
   - 🚀 **Debug Blueprint API Server** - `.env` 파일 자동 로드
   - 🔧 **Debug Blueprint Server (Manual Env)** - 수동 환경변수 설정

3. **F5** 또는 **Start Debugging** 클릭

### 방법 2: 커맨드 팔레트 사용

1. **Ctrl+Shift+P** (Mac: **Cmd+Shift+P**)
2. `Debug: Start Debugging` 입력
3. 원하는 디버그 설정 선택

## 🔧 디버깅 설정 설명

### 1. 🚀 Debug Blueprint API Server
```json
{
  "name": "🚀 Debug Blueprint API Server",
  "envFile": "${workspaceFolder}/.env",
  "env": {
    "GIN_MODE": "debug",
    "GO_ENV": "development"
  }
}
```
- `.env` 파일에서 환경변수 자동 로드
- **추천**: 일반적인 개발 시 사용

### 2. 🔧 Debug Blueprint Server (Manual Env)
```json
{
  "name": "🔧 Debug Blueprint Server (Manual Env)",
  "env": {
    "DB_HOST": "localhost",
    "DB_PORT": "5432",
    // ... 모든 환경변수 수동 설정
  }
}
```
- 환경변수를 코드에 직접 설정
- **사용**: `.env` 파일이 없거나 테스트 환경에서

## 🎯 브레이크포인트 설정

### 1. 중요한 디버깅 포인트

#### API 핸들러
```go
// internal/handlers/goal.go
func (h *ProjectHandler) CreateProjectWithMilestones(c *gin.Context) {
    // 여기에 브레이크포인트 설정 👈
    userID, exists := c.Get("user_id")

    var req models.CreateProjectWithMilestonesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // JSON 파싱 오류 디버깅 👈
    }
}
```

#### AI 서비스
```go
// internal/services/ai_bridge_service.go
func (s *BridgeAIService) GenerateMilestones(project models.CreateProjectRequest) {
    // AI 요청 디버깅 👈
    aiRequest := s.convertToAIRequest(project)
}
```

#### 데이터베이스
```go
// internal/handlers/goal.go
tx := database.GetDB().Begin()
if err := tx.Create(&project).Error; err != nil {
    // 데이터베이스 오류 디버깅 👈
}
```

### 2. 브레이크포인트 설정 방법
- **F9** 키 또는 줄 번호 왼쪽 클릭
- **조건부 브레이크포인트**: 우클릭 → "Add Conditional Breakpoint"
- **로그 포인트**: 우클릭 → "Add Logpoint"

## 🛠️ 유용한 디버깅 기능

### 1. 변수 감시 (Watch)
```go
// 감시할 변수들 예시
req.Title          // 요청 제목
project.ID         // 생성된 프로젝트 ID
aiResponse         // AI 응답 데이터
err                // 오류 정보
```

### 2. 호출 스택 (Call Stack)
- 함수 호출 경로 추적
- 어떤 함수에서 오류가 발생했는지 확인

### 3. 디버그 콘솔
- 런타임에 변수 값 확인
- Go 표현식 실행 가능

## 🧪 테스트 디버깅

### 단일 테스트 디버깅
1. 테스트 파일 열기
2. **F5** → "🧪 Debug Single Test" 선택
3. `TestFunctionName`을 실제 테스트 함수명으로 변경

### 전체 테스트 디버깅
1. **F5** → "📦 Debug All Tests" 선택
2. 모든 테스트 실행하며 디버깅

## 🚀 빌드 및 실행 작업

### VS Code Tasks 사용
- **Ctrl+Shift+P** → `Tasks: Run Task`
- 사용 가능한 작업들:
  - 🔨 **Build Blueprint Server** - 서버 빌드
  - 🚀 **Run Server** - 서버 실행
  - 🧪 **Run Tests** - 테스트 실행
  - 📦 **Go Mod Tidy** - 의존성 정리
  - 🔍 **Go Vet** - 코드 검사

### 터미널에서 직접 실행
```bash
# 개발 모드 실행
make dev-server

# 빌드 후 실행
make build && ./bin/server

# 테스트 실행
make test
```

## 🔍 일반적인 디버깅 시나리오

### 1. API 요청 오류 디버깅
```go
// 1. 핸들러 입구에 브레이크포인트
func (h *ProjectHandler) CreateProjectWithMilestones(c *gin.Context) {
    // 2. 요청 데이터 확인
    var req models.CreateProjectWithMilestonesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 3. JSON 파싱 오류 확인
        log.Printf("JSON Parse Error: %v", err)
    }
}
```

### 2. 데이터베이스 연결 문제
```go
// internal/database/database.go
func Connect(cfg *config.Config) error {
    // 브레이크포인트로 연결 정보 확인
    dsn := fmt.Sprintf("host=%s user=%s password=%s...", ...)
}
```

### 3. AI 서비스 오류
```go
// internal/services/ai_bridge_service.go
func (s *BridgeAIService) GenerateMilestones(...) {
    // AI 요청/응답 데이터 확인
    aiResponse, err := s.aiModel.GenerateMilestones(ctx, aiRequest)
}
```

## 📱 프론트엔드와 함께 디버깅

### 1. 동시 실행
```bash
# Terminal 1: 백엔드 디버깅
# VS Code F5로 디버깅 모드 실행

# Terminal 2: 프론트엔드 실행
cd web && npm run dev
```

### 2. API 테스트
- **Thunder Client** 또는 **REST Client** 확장 사용
- 브레이크포인트 설정 후 API 호출
- 요청/응답 데이터 실시간 확인

## 🎛️ 고급 디버깅 설정

### Delve 설정
```json
"go.delveConfig": {
  "dlvLoadConfig": {
    "maxStringLen": 64,
    "maxArrayValues": 64,
    "maxStructFields": -1
  },
  "showGlobalVariables": true
}
```

### 환경별 디버깅
- **Development**: 전체 로그, 상세 오류
- **Production**: 최소 로그, 보안 강화
- **Test**: 목 데이터, 빠른 실행

## 🆘 문제 해결

### 일반적인 문제들

1. **디버거가 시작되지 않음**
   - Go 확장 프로그램 설치 확인
   - `go.mod` 파일 존재 확인
   - GOPATH/GOROOT 설정 확인

2. **브레이크포인트가 작동하지 않음**
   - 코드가 실제로 실행되는지 확인
   - 컴파일러 최적화로 인한 제거 가능성
   - 디버그 모드 빌드 확인

3. **환경변수가 로드되지 않음**
   - `.env` 파일 경로 확인
   - 파일 권한 확인
   - 환경변수 이름 대소문자 확인

## 📚 참고 자료

- [Go in VS Code](https://code.visualstudio.com/docs/languages/go)
- [Delve Debugger](https://github.com/go-delve/delve)
- [Gin Framework Debug Mode](https://gin-gonic.com/docs/development/)

---

**Happy Debugging! 🐛→🎯**
