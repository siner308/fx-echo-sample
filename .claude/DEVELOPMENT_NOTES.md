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

### 2. 순환 의존성
- auth service가 user service 참조하려다 순환 의존성 발생
- auth service에서 repository 직접 사용으로 해결

### 3. 라우트 등록 복잡성
- 초기에는 그룹별 미들웨어 적용을 위한 복잡한 registrar 패턴
- 사용자 피드백으로 API별 미들웨어 선택 방식으로 단순화

## 앞으로의 개선 방향

### 1. 테스트 코드 추가
- 각 모듈별 단위 테스트
- 통합 테스트

### 2. API 문서화
- OpenAPI/Swagger 문서 자동 생성
- 엔드포인트별 상세 문서

### 3. 로깅 개선
- 구조화된 로깅
- 요청 추적을 위한 correlation ID

### 4. 에러 처리 표준화
- 커스텀 에러 타입
- 일관된 에러 응답 형식