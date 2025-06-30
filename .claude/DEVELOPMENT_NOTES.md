# 개발 노트

Claude Code를 사용한 개발 과정에서의 주요 의사결정과 학습 내용

## 아키텍처 진화

### 1. 초기 구조 → 인증 시스템 도입
- 단순한 CRUD API에서 시작
- JWT 기반 인증 시스템 추가
- 다중 토큰 타입 (access, refresh, admin) 도입

### 2. 복잡한 라우트 등록 → 유연한 미들웨어 선택
**Before:**
```go
// 복잡한 registrar 인터페이스
type AdminRouteRegistrar interface { ... }
type ProtectedRouteRegistrar interface { ... }
```

**After:**
```go
// API별 미들웨어 직접 선택
coupons.POST("", handler.CreateCoupon, adminMiddleware.VerifyAdminToken())
coupons.POST("/use", handler.UseCoupon, userMiddleware.VerifyAccessToken())
```

### 3. 미들웨어 네이밍 통일성
**Before:**
- `AdminMiddleware()`
- `JWTMiddleware()`

**After:**
- `VerifyAdminToken()`
- `VerifyAccessToken()`

## 핵심 학습 사항

### 1. Uber FX 의존성 주입 패턴
```go
// fx.In을 활용한 깔끔한 의존성 주입
type ServiceParam struct {
    fx.In
    Repository    Repository
    AccessToken   jwt.Service `name:"access"`
    RefreshToken  jwt.Service `name:"refresh"`
    Logger        *zap.Logger
}
```

### 2. 환경변수 보안
**문제:** 기본값이 있으면 보안 위험
```go
// BAD: 기본값 제공
jwtConfig := Config{
    Secret: getEnv("JWT_SECRET", "default-secret"), // 위험!
}

// GOOD: 필수값 강제
jwtConfig := Config{
    Secret: getRequiredEnv("JWT_SECRET"), // 없으면 패닉
}
```

### 3. 타입 안전 Context 키
```go
// 문자열 리터럴 에러 방지
type UserContextKey string
const UserIDKey UserContextKey = "user_id"

// 사용 시
echoCtx.Set(string(UserIDKey), userID)
```

### 4. 비즈니스 로직 중심 권한 설계
- 기술적 구분보다 비즈니스 요구사항에 맞는 권한 분리
- 관리자: 쿠폰 생성/관리, 사용자 관리
- 사용자: 쿠폰 사용, 본인 계정 관리

## 개발 과정에서의 이슈

### 1. GOPATH 설정 문제
- IDE 자동완성 오류
- GOPATH 조정으로 해결

### 2. 순환 의존성 해결 진화
**1차 시도:** auth service에서 repository 직접 사용
**2차 시도:** auth 로직을 user 모듈에 포함 → 사용자 피드백으로 철회
**최종 해결:** 의존성 역전 패턴 적용
```go
// 인터페이스 정의로 의존성 분리
type PasswordVerifier interface {
    VerifyUserPassword(email, password string) (UserInfo, error)
}

// 어댑터 패턴으로 연결
type AuthAdapter struct {
    userService user.Service
}
```

### 3. 라우트 등록 복잡성
- 초기에는 그룹별 미들웨어 적용을 위한 복잡한 registrar 패턴
- 사용자 피드백으로 API별 미들웨어 선택 방식으로 단순화

### 4. 아이템 보상 시스템 구축
- 사용자 요청으로 쿠폰 및 결제를 통한 아이템 지급 시스템 개발
- Item, Payment, Reward 모듈 추가
- 모듈 간 의존관계 최적화: item → payment/reward → coupon

### 5. 보안 강화 (패스워드)
**초기 문제:** 평문 패스워드 저장으로 코드 리뷰에서 발견된 보안 취약점
**해결 과정:**
```go
// 평문 저장 (취약)
user.Password = request.Password

// Argon2id 해싱 적용 (보안)
hashedPassword, err := security.HashPassword(request.Password, nil)
user.Password = hashedPassword
```

**보안 기능:**
- Argon2id 알고리즘 사용
- 암호학적으로 안전한 소금 생성
- 타이밍 공격 방지
- 설정 가능한 보안 파라미터

## 완료된 개선사항

### ✅ 1. 테스트 코드 구현
- **Security Package**: Argon2id 해싱/검증 테스트 (6개 테스트)
- **User Service**: CRUD 및 패스워드 보안 테스트 (7개 테스트)  
- **Auth Service**: JWT 토큰 관리 테스트 (7개 테스트)
- **통합 테스트**: 모듈 간 연동 테스트
- **벤치마크 테스트**: 성능 검증 (~54ms/op)

### ✅ 2. 모듈 의존관계 문서화
- Mermaid 다이어그램으로 시각화
- 각 모듈의 역할과 의존성 명시
- 순환 종속성 방지 전략 문서화

## 앞으로의 개선 방향

### 1. API 문서화
- OpenAPI/Swagger 문서 자동 생성
- 엔드포인트별 상세 문서
- 아이템 타입 및 결제 상태 상세 설명

### 2. 로깅 개선
- 구조화된 로깅
- 요청 추적을 위한 correlation ID
- 보안 이벤트 로깅

### 3. 에러 처리 표준화
- 커스텀 에러 타입
- 일관된 에러 응답 형식
- 다국어 에러 메시지 지원

### 4. 데이터베이스 통합
- 메모리 기반에서 실제 DB로 전환
- 마이그레이션 스크립트
- 트랜잭션 처리