# Blueprint Frontend

React + TypeScript로 구축된 Blueprint 플랫폼의 프론트엔드입니다.

## 🛠️ 기술 스택

- **React 18**: UI 라이브러리
- **TypeScript**: 타입 안전성
- **Vite**: 빌드 도구
- **Ant Design**: UI 컴포넌트
- **React Router**: 클라이언트 사이드 라우팅
- **Zustand**: 상태 관리
- **i18next**: 다국어 지원

## 🚀 시작하기

### 1. 의존성 설치
```bash
npm install
```

### 2. 개발 서버 시작
```bash
npm run dev
```

### 3. 빌드
```bash
npm run build
```

### 4. 프리뷰
```bash
npm run preview
```

## 📁 프로젝트 구조

```
src/
├── components/        # 재사용 가능한 컴포넌트
├── pages/            # 페이지 컴포넌트
├── stores/           # Zustand 스토어
├── types/            # TypeScript 타입 정의
├── styles/           # CSS 스타일
├── utils/            # 유틸리티 함수
└── locales/          # 다국어 파일
```

## 🌐 주요 기능

- **사용자 인증**: Google OAuth + JWT
- **프로젝트 관리**: 생성, 수정, 삭제
- **실시간 거래**: P2P 베팅 시스템
- **AI 제안**: 마일스톤 자동 생성
- **다크모드**: 테마 전환 지원
- **다국어**: 한국어/영어 지원

## 🔗 API 연동

백엔드 API와 연동하기 위해 `src/services/api.ts`를 통해 통신합니다.

```typescript
// 기본 API 베이스 URL
const API_BASE_URL = 'http://localhost:3000/api/v1'
```

## 📱 반응형 디자인

모든 페이지는 모바일, 태블릿, 데스크톱에서 최적화되어 작동합니다.

## 🎨 테마

- **라이트 모드**: 기본 테마
- **다크 모드**: 사용자 설정에 따라 자동 전환
