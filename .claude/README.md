# Claude Code Context

이 프로젝트는 Claude Code를 사용하여 개발된 Go Echo 프레임워크 기반 웹 애플리케이션입니다.

## 프로젝트 개요

Echo 프레임워크와 Uber FX 의존성 주입을 활용한 아이템 보상 시스템이 포함된 RESTful API 서버

### 주요 기능
- **JWT 다중 토큰 인증**: Access, Refresh, Admin 토큰 지원
- **Keycloak SSO**: 관리자 인증을 위한 SSO 통합  
- **유연한 미들웨어**: API별 개별 미들웨어 선택 가능
- **모듈식 아키텍처**: 사용자, 쿠폰, 아이템, 결제, 리워드 모듈 분리
- **아이템 보상 시스템**: 쿠폰 사용 및 결제를 통한 아이템 지급
- **Argon2id 보안**: 강력한 비밀번호 해싱 시스템
- **의존성 역전**: 순환 종속성 해결을 위한 인터페이스 기반 설계

### 아키텍처 특징
- **Uber FX**: 의존성 주입과 라이프사이클 관리
- **Type-safe Context**: Echo Context 사용 시 타입 안전성 보장
- **환경변수 보안**: 기본값 없는 필수 환경변수 설정
- **비즈니스 로직 중심**: 권한 분리 (관리자는 생성/관리, 사용자는 사용)
- **Domain-Driven Design**: 명확한 도메인 경계와 모듈 분리

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

### 아이템 관리
- `POST /api/v1/items` - 아이템 생성 (관리자 인증)
- `PUT /api/v1/items/:id` - 아이템 수정 (관리자 인증)
- `DELETE /api/v1/items/:id` - 아이템 삭제 (관리자 인증)
- `GET /api/v1/items` - 아이템 목록 (관리자 인증)
- `GET /api/v1/items/:id` - 아이템 조회 (사용자 인증)
- `GET /api/v1/items/user/:userID/inventory` - 사용자 인벤토리 조회 (사용자 인증)

### 결제 관리
- `POST /api/v1/payments` - 결제 생성 (사용자 인증)
- `PUT /api/v1/payments/:id/status` - 결제 상태 변경 (관리자 인증)
- `GET /api/v1/payments/:id` - 결제 조회 (사용자 인증)
- `GET /api/v1/payments/user/:userID` - 사용자 결제 목록 (사용자 인증)

### 리워드 관리
- `POST /api/v1/rewards` - 리워드 생성 (관리자 인증)
- `PUT /api/v1/rewards/:id` - 리워드 수정 (관리자 인증)
- `DELETE /api/v1/rewards/:id` - 리워드 삭제 (관리자 인증)
- `GET /api/v1/rewards` - 리워드 목록 (관리자 인증)
- `POST /api/v1/rewards/grant` - 리워드 지급 (관리자 인증)

### 쿠폰 관리
- `POST /api/v1/coupons` - 쿠폰 생성 (관리자 인증)
- `PUT /api/v1/coupons/:id` - 쿠폰 수정 (관리자 인증)
- `DELETE /api/v1/coupons/:id` - 쿠폰 삭제 (관리자 인증)
- `GET /api/v1/coupons` - 쿠폰 목록 (관리자 인증)
- `GET /api/v1/coupons/:id` - 쿠폰 조회 (사용자 인증)
- `POST /api/v1/coupons/redeem` - 쿠폰 사용 (사용자 인증)

## 개발 히스토리

### 주요 리팩토링
1. **라우트 등록 시스템**: 복잡한 registrar 패턴에서 API별 미들웨어 선택으로 단순화
2. **미들웨어 네이밍**: 일관성 있는 `Verify{TokenType}` 패턴 적용
3. **권한 분리**: 비즈니스 로직에 맞는 세밀한 권한 제어
4. **보안 강화**: 환경변수 기본값 제거, 타입 안전 컨텍스트 키 사용
5. **아이템 보상 시스템**: 쿠폰과 결제 통합한 아이템 지급 시스템 구축
6. **패스워드 보안**: 평문 저장에서 Argon2id 해싱으로 전환
7. **순환 종속성 해결**: 인터페이스와 어댑터 패턴으로 모듈 분리

### 기술적 의사결정
- **fx.In 패턴**: 깔끔한 의존성 주입을 위한 구조체 활용
- **Named Dependency**: JWT 서비스를 토큰 타입별로 구분하여 주입
- **Echo Context 분리**: HTTP 레이어와 비즈니스 로직 분리 유지
- **Per-API Middleware**: 그룹 단위보다 API별 미들웨어가 더 유연함
- **Repository Pattern**: 메모리 기반 구현으로 시작하여 추후 DB 전환 용이
- **Dependency Inversion**: 순환 종속성 방지를 위한 인터페이스 중심 설계

### 보안 개선사항
- **Argon2id 해싱**: 업계 표준 패스워드 해싱 알고리즘 적용
- **타이밍 공격 방지**: 패스워드 검증 시 일정한 시간 소요 보장
- **소금(Salt) 생성**: 암호학적으로 안전한 무작위 소금 생성
- **설정 가능한 보안 파라미터**: 메모리, 반복, 병렬성 조정 가능

## 환경 설정

필수 환경변수는 `.env.example` 참조
- JWT 시크릿 키들 (ACCESS_TOKEN_SECRET, ADMIN_TOKEN_SECRET 등)
- Keycloak 설정 (선택사항)
- 데이터베이스 연결 정보

## 아이템 타입 정의

- **currency**: 게임 내 화폐 (골드, 다이아몬드 등)
- **equipment**: 장비 아이템 (무기, 방어구 등)
- **consumable**: 소모품 (포션, 버프 아이템 등)
- **card**: 수집형 카드 (트레이딩 카드, 캐릭터 카드 등)
- **material**: 제작/강화 재료 (원석, 부품 등)
- **ticket**: 이용권 (던전 입장권, 특별 이벤트 티켓 등)

## 결제 상태 정의

- **pending**: 결제 대기 중
- **processing**: 결제 처리 중 
- **completed**: 결제 완료 (자동으로 아이템 지급)
- **failed**: 결제 실패
- **cancelled**: 결제 취소
- **refunded**: 결제 환불

## 테스트 커버리지

- **Security Package**: Argon2id 해싱 및 검증 테스트
- **User Service**: CRUD 작업 및 패스워드 보안 테스트
- **Auth Service**: JWT 토큰 관리 및 검증 테스트
- **Integration Tests**: 모듈 간 통합 테스트
- **Benchmark Tests**: 성능 검증 테스트

## 문서

- `docs/ECHO_CONTEXT_CONVENTIONS.md`: Echo Context 사용 규칙
- `docs/ENVIRONMENT_SETUP.md`: 환경변수 설정 가이드
- `docs/FX_CONCEPTS.md`: Uber FX 핵심 개념
- `docs/DIG_CONCEPTS.md`: Dig 라이브러리 개념
- `docs/module-dependencies.md`: 모듈 의존관계 다이어그램