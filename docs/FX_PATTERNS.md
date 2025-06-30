# FX 실전 패턴

본 프로젝트에서 실제로 사용된 FX 패턴들을 구체적인 코드 예시와 함께 설명합니다.

## 1. 모듈화 패턴

### 계층적 모듈 구조
```go
// modules/auth/module.go - 최상위 인증 모듈
var Module = fx.Options(
    jwt.Providers,     // JWT 서비스 제공자들
    keycloak.Module,   // Keycloak 클라이언트
    admin.Module,      // 관리자 인증
    user.Module,       // 사용자 인증
)

// modules/auth/admin/module.go - 관리자 모듈
var Module = fx.Options(
    fx.Provide(
        NewService,     // 관리자 서비스
        NewHandler,     // 관리자 핸들러
        NewMiddleware,  // 관리자 미들웨어
        fx.Annotate(
            NewRoutes,
            fx.As(new(router.RouteRegistrar)),
            fx.ResultTags(`group:"routes"`),
        ),
    ),
)
```

### 의존성 계층 분리
```
main
├── auth.Module
│   ├── jwt.Providers (기반 서비스)
│   ├── keycloak.Module (외부 서비스)
│   ├── admin.Module (도메인 로직)
│   └── user.Module (도메인 로직)
├── user.Module (비즈니스 로직)
└── coupon.Module (비즈니스 로직)
```

## 2. Named Dependency 패턴

### 토큰 타입별 서비스 분리
```go
// pkg/jwt/providers.go
var Providers = fx.Options(
    fx.Provide(
        fx.Annotate(
            func(logger *zap.Logger) jwt.Service {
                return newTokenService("access", logger)
            },
            fx.ResultTags(`name:"access"`),
        ),
    ),
    fx.Provide(
        fx.Annotate(
            func(logger *zap.Logger) jwt.Service {
                return newTokenService("refresh", logger)
            },
            fx.ResultTags(`name:"refresh"`),
        ),
    ),
    fx.Provide(
        fx.Annotate(
            func(logger *zap.Logger) jwt.Service {
                return newTokenService("admin", logger)
            },
            fx.ResultTags(`name:"admin"`),
        ),
    ),
)
```

### Named 의존성 소비
```go
// modules/auth/user/service.go
type ServiceParam struct {
    fx.In
    Repository    repository.Repository
    AccessToken   jwt.Service `name:"access"`   // access 토큰 서비스
    RefreshToken  jwt.Service `name:"refresh"`  // refresh 토큰 서비스
    Logger        *zap.Logger
}

func NewService(p ServiceParam) Service {
    return &service{
        repository:    p.Repository,
        accessToken:   p.AccessToken,  // 특정 토큰 서비스 사용
        refreshToken:  p.RefreshToken, // 특정 토큰 서비스 사용
        logger:        p.Logger,
    }
}
```

## 3. Group Collection 패턴

### 라우트 수집 패턴
```go
// 각 모듈에서 RouteRegistrar를 그룹에 추가
// modules/user/module.go
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),    // 인터페이스로 제공
    fx.ResultTags(`group:"routes"`),      // "routes" 그룹에 추가
)

// modules/coupon/module.go
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),
    fx.ResultTags(`group:"routes"`),
)
```

### 그룹 소비 패턴
```go
// server/server.go
type ServerParam struct {
    fx.In
    Lifecycle        fx.Lifecycle
    Logger           *zap.Logger
    LoggerMiddleware *middleware.LoggerMiddleware
    ErrorMiddleware  *middleware.ErrorMiddleware
    RouteRegistrars  []router.RouteRegistrar `group:"routes"`  // 모든 routes 수집
}

func NewEchoServer(p ServerParam) *EchoServer {
    e := echo.New()
    
    // 수집된 모든 RouteRegistrar 실행
    for _, registrar := range p.RouteRegistrars {
        registrar.RegisterRoutes(e)
    }
    
    return &EchoServer{echo: e, log: p.Logger}
}
```

## 4. Interface Abstraction 패턴

### 공통 인터페이스 정의
```go
// pkg/router/interface.go
type RouteRegistrar interface {
    RegisterRoutes(e *echo.Echo)
}
```

### 구현체별 특화
```go
// modules/user/routes.go
type Routes struct {
    handler         *Handler
    userMiddleware  *userauth.Middleware
    adminMiddleware *adminauth.Middleware
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
    api := e.Group("/api/v1")
    users := api.Group("/users")
    
    // 공개 라우트
    users.POST("/signup", r.handler.CreateUser)
    
    // 사용자 전용 라우트
    users.GET("/:id", r.handler.GetUser, r.userMiddleware.VerifyAccessToken())
    users.PUT("/:id", r.handler.UpdateUser, r.userMiddleware.VerifyAccessToken())
    users.DELETE("/:id", r.handler.DeleteUser, r.userMiddleware.VerifyAccessToken())
    
    // 관리자 전용 라우트
    users.GET("", r.handler.ListUsers, r.adminMiddleware.VerifyAdminToken())
}
```

## 5. Lifecycle 관리 패턴

### 서버 라이프사이클
```go
// server/server.go
func NewEchoServer(p ServerParam) *EchoServer {
    server := &EchoServer{
        echo: e,
        log:  p.Logger,
    }

    // 라이프사이클 훅 등록
    p.Lifecycle.Append(fx.Hook{
        OnStart: server.Start,
        OnStop:  server.Stop,
    })

    return server
}

func (s *EchoServer) Start(ctx context.Context) error {
    s.log.Info("Starting HTTP server", zap.String("addr", ":8080"))
    go func() {
        if err := s.echo.Start(":8080"); err != nil && err != http.ErrServerClosed {
            s.log.Fatal("Server failed to start", zap.Error(err))
        }
    }()
    return nil
}

func (s *EchoServer) Stop(ctx context.Context) error {
    s.log.Info("Stopping HTTP server")
    return s.echo.Shutdown(ctx)
}
```

## 6. 환경별 설정 패턴

### 조건부 Provider
```go
// pkg/keycloak/provider.go
func NewKeycloakClient(logger *zap.Logger) (*Client, error) {
    baseURL := os.Getenv("KEYCLOAK_BASE_URL")
    if baseURL == "" {
        logger.Warn("Keycloak not configured, admin SSO will be unavailable")
        return nil, nil  // nil 반환으로 선택적 의존성 처리
    }
    
    return &Client{
        baseURL:      baseURL,
        realm:        getRequiredEnv("KEYCLOAK_REALM"),
        clientID:     getRequiredEnv("KEYCLOAK_CLIENT_ID"),
        clientSecret: getRequiredEnv("KEYCLOAK_CLIENT_SECRET"),
        httpClient:   &http.Client{Timeout: 30 * time.Second},
        logger:       logger,
    }, nil
}
```

### Optional 의존성 활용
```go
// modules/auth/admin/service.go
type ServiceParam struct {
    fx.In
    Repository      repository.Repository
    AdminToken      jwt.Service `name:"admin"`
    KeycloakClient  *keycloak.Client `optional:"true"`  // 선택적 의존성
    Logger          *zap.Logger
}

func NewService(p ServiceParam) Service {
    return &service{
        repository:      p.Repository,
        adminToken:      p.AdminToken,
        keycloakClient:  p.KeycloakClient,  // nil일 수 있음
        logger:          p.Logger,
    }
}
```

## 7. 타입 안전성 패턴

### 강타입 설정 구조체
```go
// pkg/jwt/service.go
type Config struct {
    Secret    string
    ExpiresIn time.Duration
    Issuer    string
    TokenType string
}

// 환경변수를 강타입으로 변환
func newTokenService(tokenType string, logger *zap.Logger) Service {
    config := tokenConfigs[tokenType]
    secretEnvKey := strings.ToUpper(tokenType) + "_TOKEN_SECRET"
    
    jwtConfig := Config{
        Secret:    getRequiredEnv(secretEnvKey),  // 필수 환경변수
        ExpiresIn: parseExpirationTime(...),      // 타입 변환
        Issuer:    getRequiredEnv("JWT_ISSUER"),
        TokenType: tokenType,
    }
    
    return NewService(jwtConfig, logger)
}
```

## 8. 에러 처리 패턴

### Provider 에러 처리
```go
// 환경변수 검증을 Provider 생성 시점에 수행
func getRequiredEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        panic("Required environment variable not set: " + key)
    }
    return value
}

// 설정 검증
func validateJWTConfig(config Config) error {
    if config.Secret == "" {
        return fmt.Errorf("JWT secret is required")
    }
    if config.ExpiresIn <= 0 {
        return fmt.Errorf("JWT expiration time must be positive")
    }
    return nil
}
```

## 9. 테스트 친화적 패턴

### Mock Provider
```go
// 테스트용 Provider 세트
var TestProviders = fx.Options(
    fx.Provide(
        func() *zap.Logger {
            return zap.NewNop()  // 테스트용 무음 로거
        },
    ),
    fx.Provide(
        fx.Annotate(
            func() jwt.Service {
                return &mockJWTService{}  // Mock JWT 서비스
            },
            fx.ResultTags(`name:"access"`),
        ),
    ),
)
```

## 10. 성능 최적화 패턴

### Singleton 보장
FX는 기본적으로 Singleton을 보장하지만, 명시적으로 관리할 때:

```go
var (
    jwtServiceOnce sync.Once
    jwtService     jwt.Service
)

func NewJWTService() jwt.Service {
    jwtServiceOnce.Do(func() {
        jwtService = &service{...}
    })
    return jwtService
}
```

### Lazy Loading
```go
// 실제 필요할 때만 초기화되는 패턴
type LazyService struct {
    once     sync.Once
    service  Service
    factory  func() Service
}

func (l *LazyService) Get() Service {
    l.once.Do(func() {
        l.service = l.factory()
    })
    return l.service
}
```

이러한 패턴들을 통해 확장 가능하고 유지보수하기 쉬운 애플리케이션 구조를 만들 수 있습니다.