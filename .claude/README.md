# Claude Code Context

이 프로젝트는 Claude Code를 사용하여 개발된 Go Echo 프레임워크 기반 웹 애플리케이션입니다.

## 프로젝트 개요

Echo 프레임워크와 Uber FX 의존성 주입을 활용한 RESTful API 서버

### 주요 기능
- **JWT 다중 토큰 인증**: Access, Refresh, Admin 토큰 지원
- **Keycloak SSO**: 관리자 인증을 위한 SSO 통합
- **유연한 미들웨어**: API별 개별 미들웨어 선택 가능
- **모듈식 아키텍처**: 사용자, 쿠폰, 인증 모듈 분리

### 아키텍처 특징
- **Uber FX**: 의존성 주입과 라이프사이클 관리
- **Type-safe Context**: Echo Context 사용 시 타입 안전성 보장
- **환경변수 보안**: 기본값 없는 필수 환경변수 설정
- **비즈니스 로직 중심**: 권한 분리 (관리자는 생성/관리, 사용자는 사용)

## 인증 시스템

### 토큰 타입별 용도
- **Access Token**: 일반 사용자 API 접근
- **Refresh Token**: Access Token 갱신
- **Admin Token**: 관리자 API 접근

### 미들웨어 함수
- `VerifyAccessToken()`: 사용자 토큰 검증
- `VerifyAdminToken()`: 관리자 토큰 검증 (JWT + Keycloak 지원)

## API 엔드포인트

### 사용자 관리
- `POST /api/v1/users/signup` - 회원가입 (인증 불필요)
- `GET /api/v1/users/:id` - 사용자 정보 조회 (사용자 인증)
- `PUT /api/v1/users/:id` - 프로필 수정 (사용자 인증)
- `DELETE /api/v1/users/:id` - 계정 삭제 (사용자 인증)
- `GET /api/v1/users` - 전체 사용자 목록 (관리자 인증)

### 쿠폰 관리
- `POST /api/v1/coupons` - 쿠폰 생성 (관리자 인증)
- `PUT /api/v1/coupons/:id` - 쿠폰 수정 (관리자 인증)
- `DELETE /api/v1/coupons/:id` - 쿠폰 삭제 (관리자 인증)
- `GET /api/v1/coupons` - 쿠폰 목록 (관리자 인증)
- `GET /api/v1/coupons/:id` - 쿠폰 조회 (사용자 인증)
- `POST /api/v1/coupons/use` - 쿠폰 사용 (사용자 인증)

## 개발 히스토리

### 주요 리팩토링
1. **라우트 등록 시스템**: 복잡한 registrar 패턴에서 API별 미들웨어 선택으로 단순화
2. **미들웨어 네이밍**: 일관성 있는 `Verify{TokenType}` 패턴 적용
3. **권한 분리**: 비즈니스 로직에 맞는 세밀한 권한 제어
4. **보안 강화**: 환경변수 기본값 제거, 타입 안전 컨텍스트 키 사용

### 기술적 의사결정
- **fx.In 패턴**: 깔끔한 의존성 주입을 위한 구조체 활용
- **Named Dependency**: JWT 서비스를 토큰 타입별로 구분하여 주입
- **Echo Context 분리**: HTTP 레이어와 비즈니스 로직 분리 유지
- **Per-API Middleware**: 그룹 단위보다 API별 미들웨어가 더 유연함

## 환경 설정

필수 환경변수는 `.env.example` 참조
- JWT 시크릿 키들 (ACCESS_TOKEN_SECRET, ADMIN_TOKEN_SECRET 등)
- Keycloak 설정 (선택사항)
- 데이터베이스 연결 정보

## 문서

- `docs/ECHO_CONTEXT_CONVENTIONS.md`: Echo Context 사용 규칙
- `docs/ENVIRONMENT_SETUP.md`: 환경변수 설정 가이드