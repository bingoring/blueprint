# Google OAuth 2.0 설정 가이드

현재 제공된 파일(`blueprint-467515-134e003cd7f4.json`)은 **Service Account** 키입니다.
Google OAuth 웹 로그인을 위해서는 **OAuth 2.0 Client ID**가 필요합니다.

## 🔧 Google Cloud Console 설정

### 1. Google Cloud Console 접속
- https://console.cloud.google.com/ 접속
- 프로젝트 선택: `blueprint-467515`

### 2. OAuth 2.0 Client ID 생성
1. **API 및 서비스** > **사용자 인증 정보** 이동
2. **+ 사용자 인증 정보 만들기** 클릭
3. **OAuth 클라이언트 ID** 선택
4. 애플리케이션 유형: **웹 애플리케이션** 선택
5. 이름: `Blueprint Web App` 입력

### 3. 승인된 리디렉션 URI 추가
**승인된 리디렉션 URI**에 다음 추가:
```
http://localhost:8080/api/v1/auth/google/callback
http://localhost:3000/auth/google/callback
```

### 4. 클라이언트 ID 및 보안 비밀 복사
생성 완료 후 다음 정보 복사:
- **클라이언트 ID**: `123456789-abcdef.apps.googleusercontent.com` 형태
- **클라이언트 보안 비밀**: `GOCSPX-` 로 시작하는 문자열

## 🔄 환경변수 업데이트

### 방법 1: 스크립트 수정
`scripts/setup-env.sh` 파일에서 다음 라인 수정:
```bash
export GOOGLE_CLIENT_ID=발급받은_클라이언트_ID
export GOOGLE_CLIENT_SECRET=발급받은_클라이언트_보안_비밀
```

### 방법 2: 직접 환경변수 설정
```bash
export GOOGLE_CLIENT_ID=발급받은_클라이언트_ID
export GOOGLE_CLIENT_SECRET=발급받은_클라이언트_보안_비밀
```

## 📋 현재 상황

**프로젝트 ID**: `blueprint-467515` ✅ 확인됨
**Service Account**: 있음 (필요시 서버 간 통신 용도)
**OAuth Client ID**: ✅ **설정 완료!**
**클라이언트 ID**: `475922118539-g8plhmjifnenttr36956q7a437ols7eq.apps.googleusercontent.com`

## 🚀 테스트

OAuth 설정 완료 후:
```bash
# 환경변수 설정
source scripts/setup-env.sh

# 백엔드 시작
make run-backend

# 프론트엔드 시작 (다른 터미널)
make run-frontend
```

브라우저에서 http://localhost:3000 접속하여 Google 로그인 테스트

## 🔒 보안 참고사항

1. **클라이언트 보안 비밀**은 절대 공개 저장소에 커밋하지 마세요
2. 프로덕션에서는 HTTPS 리디렉션 URI 사용
3. 승인된 도메인 설정으로 보안 강화
