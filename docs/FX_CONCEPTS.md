# Uber FX 핵심 개념

Uber FX는 Go 애플리케이션을 위한 의존성 주입 프레임워크입니다. 이 문서는 본 프로젝트에서 사용된 FX의 핵심 개념들을 정리합니다.

## 1. Application과 Module

### Application
FX 애플리케이션은 `fx.New()`로 생성되며, 여러 Module들을 조합하여 구성됩니다.

```go
// main.go
func main() {
    fx.New(
        fx.Provide(
            // 기본 의존성들
            zap.NewProduction,
            validator.NewValidator,
            middleware.NewLoggerMiddleware,
            middleware.NewErrorMiddleware,
            server.NewEchoServer,
        ),
        // 모듈들 조합
        auth.Module,
        user.Module,
        coupon.Module,
        fx.Invoke(func(s *server.EchoServer) {
            // 애플리케이션 시작 시 실행될 함수
        }),
    ).Run()
}
```

### Module
Module은 관련된 의존성들을 그룹화하는 단위입니다.

```go
// modules/auth/module.go
var Module = fx.Options(
    admin.Module,    // 하위 모듈 포함
    user.Module,     // 하위 모듈 포함
)

// modules/user/module.go
var Module = fx.Options(
    repository.Module,
    fx.Provide(
        NewService,
        NewHandler,
        fx.Annotate(
            NewRoutes,
            fx.As(new(router.RouteRegistrar)),
            fx.ResultTags(`group:"routes"`),
        ),
    ),
)
```

## 2. Provider 패턴

### 기본 Provider
생성자 함수를 Provider로 등록하여 의존성을 제공합니다.

```go
// 기본 Provider 등록
fx.Provide(
    NewService,      // func NewService(...) *Service
    NewHandler,      // func NewHandler(...) *Handler
    NewMiddleware,   // func NewMiddleware(...) *Middleware
)
```

### Named Provider
같은 타입의 다른 인스턴스들을 구분하기 위해 이름을 부여합니다.

```go
// pkg/jwt/providers.go - Named Provider 생성
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

## 3. fx.In 패턴

### 기본 fx.In 사용
의존성을 구조체로 그룹화하여 깔끔하게 주입받습니다.

```go
// modules/auth/user/service.go
type ServiceParam struct {
    fx.In
    Repository    repository.Repository
    AccessToken   jwt.Service `name:"access"`   // Named 의존성
    RefreshToken  jwt.Service `name:"refresh"`  // Named 의존성
    Logger        *zap.Logger
}

func NewService(p ServiceParam) Service {
    return &service{
        repository:    p.Repository,
        accessToken:   p.AccessToken,
        refreshToken:  p.RefreshToken,
        logger:        p.Logger,
    }
}
```

### Named 의존성 주입
`name` 태그를 사용하여 특정 Named Provider를 주입받습니다.

```go
type ServerParam struct {
    fx.In
    Lifecycle        fx.Lifecycle
    Logger           *zap.Logger
    LoggerMiddleware *middleware.LoggerMiddleware
    ErrorMiddleware  *middleware.ErrorMiddleware
    RouteRegistrars  []router.RouteRegistrar `group:"routes"`  // Group 의존성
}
```

## 4. Group과 ResultTags

### Group Provider
여러 인스턴스를 하나의 그룹으로 수집합니다.

```go
// 각 모듈에서 RouteRegistrar를 "routes" 그룹에 추가
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),
    fx.ResultTags(`group:"routes"`),    // "routes" 그룹에 추가
),
```

### Group Consumer
그룹으로 수집된 의존성들을 슬라이스로 주입받습니다.

```go
type ServerParam struct {
    fx.In
    // ...
    RouteRegistrars []router.RouteRegistrar `group:"routes"`  // 모든 routes 그룹 수집
}

func NewEchoServer(p ServerParam) *EchoServer {
    // 모든 RouteRegistrar들을 순회하며 라우트 등록
    for _, registrar := range p.RouteRegistrars {
        registrar.RegisterRoutes(e)
    }
}
```

## 5. Interface와 fx.As

### Interface 구현 등록
구현체를 인터페이스로 제공합니다.

```go
fx.Annotate(
    NewRoutes,                           // 구현체 생성자
    fx.As(new(router.RouteRegistrar)),   // 인터페이스로 제공
    fx.ResultTags(`group:"routes"`),     // 그룹에 추가
)
```

### 실제 구현체
```go
// modules/user/routes.go
type Routes struct {
    handler         *Handler
    userMiddleware  *userauth.Middleware
    adminMiddleware *adminauth.Middleware
}

// router.RouteRegistrar 인터페이스 구현
func (r *Routes) RegisterRoutes(e *echo.Echo) {
    // 라우트 등록 로직
}
```

## 6. Lifecycle 관리

### Lifecycle Hook
애플리케이션 시작/종료 시 실행될 로직을 등록합니다.

```go
// server/server.go
func NewEchoServer(p ServerParam) *EchoServer {
    server := &EchoServer{
        echo: e,
        log:  p.Logger,
    }

    // 라이프사이클 훅 등록
    p.Lifecycle.Append(fx.Hook{
        OnStart: server.Start,    // 애플리케이션 시작 시
        OnStop:  server.Stop,     // 애플리케이션 종료 시
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

## 7. fx.Options 조합

### 모듈 조합
여러 옵션들을 하나로 묶어 재사용 가능한 모듈을 만듭니다.

```go
// modules/auth/module.go
var Module = fx.Options(
    jwt.Providers,     // JWT 서비스들
    keycloak.Module,   // Keycloak 클라이언트
    admin.Module,      // Admin 모듈
    user.Module,       // User 모듈
)
```

## 8. 모범 사례

### 1. 의존성 그룹화
```go
// ✅ 좋은 예: 관련 의존성들을 구조체로 그룹화
type ServiceParam struct {
    fx.In
    Repository   repository.Repository
    AccessToken  jwt.Service `name:"access"`
    Logger       *zap.Logger
}

// ❌ 나쁜 예: 개별 매개변수로 받기
func NewService(repo repository.Repository, token jwt.Service, logger *zap.Logger) Service
```

### 2. 명확한 네이밍
```go
// ✅ 좋은 예: 용도가 명확한 이름
`name:"access"`
`name:"refresh"`
`name:"admin"`

// ❌ 나쁜 예: 모호한 이름
`name:"jwt1"`
`name:"jwt2"`
```

### 3. 인터페이스 활용
```go
// ✅ 좋은 예: 인터페이스로 추상화
fx.As(new(router.RouteRegistrar))

// ❌ 나쁜 예: 구체 타입 직접 사용
// 구체 타입에 직접 의존하면 결합도가 높아짐
```

이러한 패턴들을 통해 FX는 타입 안전성을 보장하면서도 유연한 의존성 주입을 제공합니다.