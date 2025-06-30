# 아키텍처 개요

이 문서는 fx-echo-sample 프로젝트의 전체 아키텍처를 설명합니다.

## 시스템 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                         Client                              │
│                    (Web/Mobile App)                        │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP/HTTPS
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Echo HTTP Server                        │
│                   (Port 8080)                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Middleware Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    CORS     │ │   Logging   │ │  Recovery   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│  ┌─────────────┐ ┌─────────────┐                           │
│  │ Access Auth │ │ Admin Auth  │                           │
│  └─────────────┘ └─────────────┘                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                     Router Layer                           │
│  /api/v1/users      /api/v1/items      /api/v1/coupons     │
│  /api/v1/payments   /api/v1/rewards    /api/v1/auth        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Handler Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    User     │ │    Item     │ │   Coupon    │           │
│  │  Handlers   │ │  Handlers   │ │  Handlers   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   Payment   │ │   Reward    │ │    Auth     │           │
│  │  Handlers   │ │  Handlers   │ │  Handlers   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Service Layer                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    User     │ │    Item     │ │   Coupon    │           │
│  │  Service    │ │  Service    │ │  Service    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   Payment   │ │   Reward    │ │    Auth     │           │
│  │  Service    │ │  Service    │ │  Service    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                Repository Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    User     │ │    Item     │ │   Coupon    │           │
│  │ Repository  │ │ Repository  │ │ Repository  │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│  ┌─────────────┐ ┌─────────────┐                           │
│  │   Payment   │ │   Reward    │                           │
│  │ Repository  │ │ Repository  │                           │
│  └─────────────┘ └─────────────┘                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Data Layer                               │
│                (In-Memory Storage)                         │
│   ┌─────────────────────────────────────────────────────┐   │
│   │             Future: Database                        │   │
│   │        (PostgreSQL/MySQL/MongoDB)                  │   │
│   └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 모듈 구조

### 핵심 패키지
```
fxserver/
├── main.go                 # 애플리케이션 진입점
├── server/                 # HTTP 서버 설정
│   └── server.go
├── middleware/             # HTTP 미들웨어
│   ├── auth.go
│   ├── cors.go
│   └── error.go
├── pkg/                    # 공통 패키지
│   ├── jwt/               # JWT 토큰 관리
│   ├── keycloak/          # Keycloak 통합
│   ├── router/            # 라우터 설정
│   └── security/          # 보안 (Argon2id)
└── modules/               # 비즈니스 모듈
    ├── auth/              # 인증 서비스
    ├── user/              # 사용자 관리
    ├── item/              # 아이템 관리
    ├── payment/           # 결제 처리
    ├── reward/            # 리워드 지급
    └── coupon/            # 쿠폰 시스템
```

### 모듈별 구조
각 비즈니스 모듈은 다음과 같은 구조를 따릅니다:

```
modules/{module_name}/
├── entity/                # 데이터 모델
│   └── {entity}.go
├── repository/            # 데이터 접근 계층
│   └── repository.go
├── service.go             # 비즈니스 로직
├── handler.go             # HTTP 핸들러
├── routes.go              # 라우트 정의
├── module.go              # FX 모듈 정의
└── {module}_test.go       # 테스트 코드
```

## 의존성 주입 (Uber FX)

### FX 모듈 구조
```go
// 각 모듈은 FX 모듈로 정의됩니다
var Module = fx.Options(
    fx.Provide(
        NewService,      // 서비스 제공
        NewHandler,      // 핸들러 제공
        NewRepository,   // 리포지토리 제공
    ),
    fx.Invoke(
        RegisterRoutes,  // 라우트 등록
    ),
)
```

### 의존성 주입 패턴
```go
// fx.In을 사용한 의존성 주입
type ServiceParam struct {
    fx.In
    Repository    Repository
    Logger        *zap.Logger
    AccessToken   jwt.Service `name:"access"`
    RefreshToken  jwt.Service `name:"refresh"`
}

func NewService(param ServiceParam) Service {
    return &service{
        repo:         param.Repository,
        logger:       param.Logger,
        accessToken:  param.AccessToken,
        refreshToken: param.RefreshToken,
    }
}
```

## 인증 및 인가

### 다중 토큰 시스템
```
┌─────────────────────────────────────────────────────────────┐
│                     Token Types                            │
├─────────────────────────────────────────────────────────────┤
│  Access Token    │ 사용자 API 접근 (1시간)                  │
│  Refresh Token   │ Access Token 갱신 (24시간)               │
│  Admin Token     │ 관리자 API 접근 (JWT/Keycloak)           │
└─────────────────────────────────────────────────────────────┘
```

### 미들웨어 선택 방식
```go
// API별 개별 미들웨어 선택
users.GET("/:id", handler.GetUser, middleware.VerifyAccessToken())
users.GET("", handler.ListUsers, middleware.VerifyAdminToken())
```

## 데이터 흐름

### 1. 사용자 등록 플로우
```
Client Request → Handler → Service → Repository → Response
     ↓
Password → Argon2id Hashing → Secure Storage
```

### 2. 인증 플로우
```
Login Request → Auth Service → Password Verification → JWT Generation
     ↓
User Service ← Password Verifier Interface ← Auth Adapter
```

### 3. 아이템 지급 플로우
```
Payment/Coupon → Reward Service → Item Service → User Inventory Update
     ↓
Payment Status: completed → Auto Item Grant
```

## 보안 아키텍처

### 패스워드 보안
```go
// Argon2id 설정
type PasswordConfig struct {
    Memory      uint32 // 메모리 사용량 (KB)
    Iterations  uint32 // 반복 횟수
    Parallelism uint8  // 병렬성
    SaltLength  uint32 // 소금 길이
    KeyLength   uint32 // 키 길이
}
```

### 환경변수 보안
```go
// 기본값 없는 필수 환경변수
func getRequiredEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        panic(fmt.Sprintf("Required environment variable %s is not set", key))
    }
    return value
}
```

## 에러 처리

### 계층별 에러 처리
```
Repository Layer → Service Layer → Handler Layer → Client
     │                  │               │
Custom Errors    Business Logic    HTTP Status
                     Errors         Codes
```

### 표준 에러 응답
```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": "Additional details"
}
```

## 성능 고려사항

### 메모리 기반 저장소
- 현재: In-memory storage (개발/테스트용)
- 향후: Database integration (PostgreSQL/MySQL)

### 동시성 처리
- JWT 토큰 생성/검증은 상태 없음 (stateless)
- Repository 레벨에서 동시성 제어 필요

### 확장성
- 모듈식 설계로 수평 확장 가능
- 각 서비스는 독립적으로 배포 가능한 구조

## 테스트 전략

### 테스트 계층
```
Unit Tests       → Service/Repository 레벨
Integration Tests → Module 간 통합
End-to-End Tests → 전체 API 플로우
Benchmark Tests  → 성능 검증
```

### Mock 기반 테스트
```go
// testify/mock을 사용한 의존성 모킹
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(user *entity.User) error {
    args := m.Called(user)
    return args.Error(0)
}
```

## 향후 아키텍처 개선 방향

### 1. 마이크로서비스 분리
- 각 모듈을 독립 서비스로 분리
- gRPC 기반 서비스 간 통신

### 2. 이벤트 기반 아키텍처
- 결제 완료 → 아이템 지급 이벤트
- NATS/Kafka 기반 메시징

### 3. 데이터베이스 분리
- 모듈별 전용 데이터베이스
- CQRS 패턴 적용

### 4. 캐싱 레이어
- Redis 기반 세션 관리
- 자주 조회되는 데이터 캐싱

### 5. API Gateway
- 인증/인가 중앙화
- 라우팅 및 로드 밸런싱