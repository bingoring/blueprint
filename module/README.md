# Module Directory

MSA (Microservice Architecture) 대비 공통 모듈 디렉토리입니다.

## 구조

```
module/
├── common/          # 공통 유틸리티
├── types/           # 공통 타입 정의
├── proto/           # gRPC 프로토콜 정의
├── config/          # 공통 설정
├── database/        # 데이터베이스 공통 모듈
├── auth/            # 인증/권한 공통 모듈
├── logging/         # 로깅 공통 모듈
└── monitoring/      # 모니터링 공통 모듈
```

## 사용법

각 마이크로서비스에서 필요한 공통 모듈을 import하여 사용합니다.

```go
import (
    "blueprint/module/common"
    "blueprint/module/types"
    "blueprint/module/auth"
)
```

## 예정된 마이크로서비스

- **User Service**: 사용자 관리
- **Project Service**: 프로젝트 관리
- **Trading Service**: 거래 엔진
- **Notification Service**: 알림 서비스
- **Analytics Service**: 분석 서비스
