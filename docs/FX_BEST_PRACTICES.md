# FX 모범 사례

본 프로젝트 개발 과정에서 얻은 FX 사용 시 모범 사례와 피해야 할 안티패턴들을 정리합니다.

## 1. 모듈 설계 원칙

### ✅ 도메인별 모듈 분리
```go
// 좋은 예: 도메인별로 명확하게 분리
var Module = fx.Options(
    auth.Module,      // 인증 도메인
    user.Module,      // 사용자 도메인  
    coupon.Module,    // 쿠폰 도메인
)
```

### ❌ 기술적 계층별 분리 (안티패턴)
```go
// 나쁜 예: 기술적 계층으로만 분리
var Module = fx.Options(
    handler.Module,    // 모든 핸들러
    service.Module,    // 모든 서비스
    repository.Module, // 모든 리포지토리
)
```

### ✅ 계층적 의존성 구조
```
main
├── 기반 서비스 (Logger, Validator, Database)
├── 공통 패키지 (JWT, Keycloak)
├── 도메인 서비스 (Auth, User, Coupon)
└── HTTP 서버 (Server, Middleware)
```

## 2. 의존성 주입 모범 사례

### ✅ fx.In 구조체 활용
```go
// 좋은 예: 관련 의존성들을 구조체로 그룹화
type ServiceParam struct {
    fx.In
    Repository    repository.Repository
    AccessToken   jwt.Service `name:"access"`
    RefreshToken  jwt.Service `name:"refresh"`
    Logger        *zap.Logger
}

func NewService(p ServiceParam) Service {
    return &service{
        repository:   p.Repository,
        accessToken:  p.AccessToken,
        refreshToken: p.RefreshToken,
        logger:       p.Logger,
    }
}
```

### ❌ 개별 매개변수 (안티패턴)
```go
// 나쁜 예: 매개변수가 많아지면 관리가 어려움
func NewService(
    repo repository.Repository,
    accessToken jwt.Service,
    refreshToken jwt.Service,
    logger *zap.Logger,
    config *Config,
    metrics *Metrics,
) Service {
    // 매개변수가 너무 많음
}
```

### ✅ 명확한 Named 의존성
```go
// 좋은 예: 용도가 명확한 이름
fx.ResultTags(`name:"access"`),   // 접근 토큰
fx.ResultTags(`name:"refresh"`),  // 갱신 토큰
fx.ResultTags(`name:"admin"`),    // 관리자 토큰
```

### ❌ 모호한 Named 의존성
```go
// 나쁜 예: 용도를 알 수 없는 이름
fx.ResultTags(`name:"jwt1"`),     // 무엇인지 불분명
fx.ResultTags(`name:"token2"`),   // 용도 불명확  
fx.ResultTags(`name:"service"`),  // 너무 일반적
```

## 3. Provider 설계 원칙

### ✅ 단일 책임 Provider
```go
// 좋은 예: 각 Provider가 하나의 명확한 역할
func NewAccessTokenService(logger *zap.Logger) jwt.Service {
    return newTokenService("access", logger)
}

func NewRefreshTokenService(logger *zap.Logger) jwt.Service {
    return newTokenService("refresh", logger)
}
```

### ❌ 다중 책임 Provider
```go
// 나쁜 예: 하나의 Provider가 너무 많은 일을 함
func NewAllServices() (*UserService, *AuthService, *CouponService, error) {
    // 너무 많은 서비스를 한 번에 생성
    // 의존성 추적이 어려움
}
```

### ✅ 환경변수 검증
```go
// 좋은 예: 필수 환경변수 검증
func getRequiredEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        panic("Required environment variable not set: " + key)
    }
    return value
}

func newTokenService(tokenType string, logger *zap.Logger) Service {
    secretEnvKey := strings.ToUpper(tokenType) + "_TOKEN_SECRET"
    jwtConfig := Config{
        Secret: getRequiredEnv(secretEnvKey), // 필수 값 검증
        // ...
    }
    return NewService(jwtConfig, logger)
}
```

### ❌ 기본값 의존 (보안 위험)
```go
// 나쁜 예: 보안에 민감한 값에 기본값 제공
func NewJWTService() jwt.Service {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "default-secret" // 보안 위험!
    }
    // ...
}
```

## 4. 인터페이스 활용

### ✅ 적절한 추상화
```go
// 좋은 예: 의미 있는 인터페이스 정의
type RouteRegistrar interface {
    RegisterRoutes(e *echo.Echo)
}

// 구현체를 인터페이스로 제공
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),
    fx.ResultTags(`group:"routes"`),
)
```

### ❌ 과도한 추상화
```go
// 나쁜 예: 불필요한 인터페이스
type StringProvider interface {
    GetString() string
}

type IntProvider interface {
    GetInt() int
}
// 단순한 값들까지 인터페이스로 추상화할 필요 없음
```

### ✅ 테스트 친화적 인터페이스
```go
// 좋은 예: Mock하기 쉬운 인터페이스
type JWTService interface {
    GenerateToken(userID int) (string, error)
    ValidateToken(token string) (*Claims, error)
}

// 테스트에서 쉽게 Mock 가능
type MockJWTService struct{}
func (m *MockJWTService) GenerateToken(userID int) (string, error) {
    return "mock-token", nil
}
```

## 5. 에러 처리 전략

### ✅ 실패 빠른 검증 (Fail Fast)
```go
// 좋은 예: 애플리케이션 시작 시점에 설정 검증
func NewJWTService(tokenType string, logger *zap.Logger) Service {
    config, exists := tokenConfigs[tokenType]
    if !exists {
        logger.Error("Unknown token type", zap.String("type", tokenType))
        panic("Unknown token type: " + tokenType) // 즉시 실패
    }
    
    jwtConfig := Config{
        Secret: getRequiredEnv(secretEnvKey), // 환경변수 검증
        // ...
    }
    
    return NewService(jwtConfig, logger)
}
```

### ❌ 지연된 에러 처리
```go
// 나쁜 예: 런타임에야 에러 발견
func NewJWTService() Service {
    return &service{} // 설정 검증 없음
}

func (s *service) GenerateToken(userID int) (string, error) {
    if s.secret == "" {
        return "", errors.New("JWT secret not configured") // 너무 늦은 발견
    }
    // ...
}
```

## 6. 라이프사이클 관리

### ✅ 적절한 라이프사이클 훅
```go
// 좋은 예: 리소스 정리까지 고려
func NewEchoServer(p ServerParam) *EchoServer {
    server := &EchoServer{echo: e, log: p.Logger}

    p.Lifecycle.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            server.log.Info("Starting HTTP server")
            go server.echo.Start(":8080")
            return nil
        },
        OnStop: func(ctx context.Context) error {
            server.log.Info("Stopping HTTP server")
            return server.echo.Shutdown(ctx) // Graceful shutdown
        },
    })

    return server
}
```

### ❌ 리소스 누수
```go
// 나쁜 예: 정리 로직 없음
func NewServer(p ServerParam) *Server {
    server := &Server{}
    
    p.Lifecycle.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go server.Start() // 시작만 하고 정리 로직 없음
            return nil
        },
        // OnStop 누락 - 리소스 누수 가능
    })
    
    return server
}
```

## 7. 테스트 지원

### ✅ 테스트 친화적 설계
```go
// 좋은 예: 테스트용 모듈 별도 제공
var TestModule = fx.Options(
    fx.Provide(
        func() *zap.Logger {
            return zap.NewNop() // 테스트용 무음 로거
        },
    ),
    fx.Provide(
        fx.Annotate(
            func() jwt.Service {
                return &MockJWTService{} // Mock 서비스
            },
            fx.ResultTags(`name:"access"`),
        ),
    ),
)

// 테스트에서 사용
func TestUserService(t *testing.T) {
    app := fx.New(
        TestModule,
        user.Module,
        fx.Invoke(func(service user.Service) {
            // 테스트 로직
        }),
    )
    app.Start(context.Background())
    defer app.Stop(context.Background())
}
```

### ❌ 테스트 어려운 설계
```go
// 나쁜 예: 하드코딩된 의존성
func NewService() Service {
    db := sql.Open("postgres", "production-db-url") // 하드코딩
    logger := zap.NewProduction()                   // 프로덕션 로거
    return &service{db: db, logger: logger}
}
```

## 8. 성능 고려사항

### ✅ 지연 초기화
```go
// 좋은 예: 비용이 큰 리소스는 실제 사용 시점에 초기화
type KeycloakClient struct {
    config     Config
    httpClient *http.Client
    once       sync.Once
    client     *keycloak.Client
}

func (k *KeycloakClient) GetClient() *keycloak.Client {
    k.once.Do(func() {
        k.client = keycloak.New(k.config) // 실제 사용 시점에 초기화
    })
    return k.client
}
```

### ❌ 불필요한 사전 초기화
```go
// 나쁜 예: 사용하지도 않는 리소스를 미리 초기화
func NewAllClients() (*DBClient, *RedisClient, *S3Client, *EmailClient) {
    // 모든 클라이언트를 무조건 초기화 (비효율적)
    return db, redis, s3, email
}
```

## 9. 문서화와 유지보수

### ✅ 명확한 모듈 문서화
```go
// 좋은 예: 모듈의 역할과 의존성을 명확히 문서화
// Package auth provides authentication and authorization services.
// 
// This module includes:
// - JWT token services (access, refresh, admin)
// - Keycloak SSO integration
// - Authentication middleware
//
// Dependencies:
// - zap.Logger for logging
// - Database for user storage
// - Environment variables for configuration
package auth

var Module = fx.Options(
    jwt.Providers,
    keycloak.Module,
    admin.Module,
    user.Module,
)
```

### ✅ 의존성 그래프 시각화
```go
// 개발 시 의존성 확인을 위한 유틸리티
func visualizeDependencies() {
    app := fx.New(
        // 모든 모듈 포함
        auth.Module,
        user.Module,
        coupon.Module,
        fx.Invoke(func() {}), // 빈 invoke
    )
    
    if err := app.Err(); err != nil {
        log.Fatal("Dependency error:", err)
    }
}
```

## 10. 마이그레이션 전략

### ✅ 점진적 FX 도입
```go
// 1단계: 기존 코드 유지하면서 FX Provider 추가
func NewLegacyService() *Service {
    // 기존 생성 로직 유지
    return &Service{}
}

// 2단계: FX 패턴으로 점진적 변환
type ServiceParam struct {
    fx.In
    // 새로운 의존성들 추가
}

func NewService(p ServiceParam) Service {
    // FX 패턴으로 변환
}
```

이러한 모범 사례들을 따르면 유지보수하기 쉽고 확장 가능한 FX 애플리케이션을 구축할 수 있습니다.