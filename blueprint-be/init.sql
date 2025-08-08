-- Blueprint 데이터베이스 초기화 스크립트

-- 필요한 확장 설치
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- 데이터베이스 설정
ALTER DATABASE blueprint SET timezone TO 'UTC';

-- 인덱스 성능 향상을 위한 설정
-- (GORM이 자동으로 생성하는 테이블들에 추가적인 인덱스가 필요한 경우 여기에 추가)

-- 초기 데이터 삽입 (필요한 경우)
-- 예: 기본 관리자 계정, 카테고리 설정 등

-- 로그 출력
\echo 'Blueprint database initialized successfully!'
