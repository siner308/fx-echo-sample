# Dig 핵심 개념

Dig는 Uber FX의 기반이 되는 의존성 주입 컨테이너입니다. FX는 Dig를 래핑하여 더 높은 수준의 추상화를 제공합니다.

## 1. Container와 Provider

### Container
Dig의 핵심은 의존성을 관리하는 Container입니다.

```go
import "go.uber.org/dig"

// 기본 Container 생성
container := dig.New()

// Provider 등록
container.Provide(func() *Database {
    return &Database{}
})

// 의존성 해결 및 실행
container.Invoke(func(db *Database) {
    // Database 사용
})
```

### Provider 함수
생성자 함수를 Provider로 등록합니다.

```go
// 간단한 Provider
func NewDatabase() *Database {
    return &Database{}
}

// 의존성이 있는 Provider
func NewUserService(db *Database) *UserService {
    return &UserService{db: db}
}

// Container에 등록
container.Provide(NewDatabase)
container.Provide(NewUserService)
```

## 2. dig.In 구조체

### 기본 dig.In 사용
여러 의존성을 구조체로 그룹화하여 받습니다.

```go
// 의존성 그룹화
type ServiceDeps struct {
    dig.In
    Database *Database
    Logger   *Logger
    Config   *Config
}

func NewService(deps ServiceDeps) *Service {
    return &Service{
        db:     deps.Database,
        logger: deps.Logger,
        config: deps.Config,
    }
}
```

### 본 프로젝트에서의 활용
```go
// modules/auth/user/service.go
type ServiceParam struct {
    fx.In  // fx.In은 dig.In을 래핑
    Repository    repository.Repository
    AccessToken   jwt.Service `name:"access"`
    RefreshToken  jwt.Service `name:"refresh"`
    Logger        *zap.Logger
}
```

## 3. dig.Out 구조체

### 여러 값 반환
하나의 Provider에서 여러 의존성을 제공합니다.

```go
type DatabaseResult struct {
    dig.Out
    Database *Database
    Migrator *Migrator
}

func NewDatabase() DatabaseResult {
    db := &Database{}
    return DatabaseResult{
        Database: db,
        Migrator: &Migrator{db: db},
    }
}
```

### Named Output
같은 타입의 서로 다른 인스턴스를 제공합니다.

```go
type JWTResult struct {
    dig.Out
    AccessToken  jwt.Service `name:"access"`
    RefreshToken jwt.Service `name:"refresh"`
}

func NewJWTServices() JWTResult {
    return JWTResult{
        AccessToken:  jwt.NewService("access_secret"),
        RefreshToken: jwt.NewService("refresh_secret"),
    }
}
```

## 4. Named Dependencies

### Named Provider
이름으로 구분되는 의존성을 제공합니다.

```go
// 본 프로젝트의 JWT Provider 패턴
func provideAccessTokenService(logger *zap.Logger) jwt.Service {
    return newTokenService("access", logger)
}

func provideRefreshTokenService(logger *zap.Logger) jwt.Service {
    return newTokenService("refresh", logger)
}

// Named로 등록 (FX 문법)
fx.Annotate(
    provideAccessTokenService,
    fx.ResultTags(`name:"access"`),
)
```

### Named Consumer
이름으로 특정 의존성을 주입받습니다.

```go
type ServiceParam struct {
    dig.In
    AccessToken  jwt.Service `name:"access"`   // "access" 이름의 jwt.Service
    RefreshToken jwt.Service `name:"refresh"`  // "refresh" 이름의 jwt.Service
}
```

## 5. Group Dependencies

### Group Provider
여러 인스턴스를 그룹으로 수집합니다.

```go
// 본 프로젝트의 RouteRegistrar 그룹 패턴
type RoutesResult struct {
    dig.Out
    RouteRegistrar router.RouteRegistrar `group:"routes"`
}

func NewUserRoutes(...) RoutesResult {
    return RoutesResult{
        RouteRegistrar: &UserRoutes{...},
    }
}
```

### Group Consumer
그룹의 모든 인스턴스를 슬라이스로 받습니다.

```go
type ServerDeps struct {
    dig.In
    RouteRegistrars []router.RouteRegistrar `group:"routes"`
}

func NewServer(deps ServerDeps) *Server {
    for _, registrar := range deps.RouteRegistrars {
        registrar.RegisterRoutes(server.engine)
    }
    return server
}
```

## 6. Optional Dependencies

### Optional 의존성
없어도 되는 의존성을 처리합니다.

```go
type ServiceDeps struct {
    dig.In
    Database *Database
    Cache    *Cache `optional:"true"`  // 없어도 됨
}

func NewService(deps ServiceDeps) *Service {
    service := &Service{db: deps.Database}
    
    if deps.Cache != nil {
        service.cache = deps.Cache
    }
    
    return service
}
```

## 7. Interface Binding

### 인터페이스로 제공
구현체를 인터페이스로 바인딩합니다.

```go
// 인터페이스 정의
type UserRepository interface {
    GetUser(id int) (*User, error)
}

// 구현체
type PostgresUserRepository struct{}

func (r *PostgresUserRepository) GetUser(id int) (*User, error) {
    // 구현
}

// 인터페이스로 바인딩하여 제공
func NewUserRepository() UserRepository {
    return &PostgresUserRepository{}
}
```

### 본 프로젝트에서의 활용
```go
// pkg/router/interface.go
type RouteRegistrar interface {
    RegisterRoutes(e *echo.Echo)
}

// modules/user/routes.go - 인터페이스 구현
func (r *Routes) RegisterRoutes(e *echo.Echo) {
    // 라우트 등록
}

// modules/user/module.go - 인터페이스로 제공
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),  // 인터페이스로 바인딩
    fx.ResultTags(`group:"routes"`),
)
```

## 8. 에러 처리

### 순환 의존성
Dig는 순환 의존성을 자동으로 감지합니다.

```go
// ❌ 순환 의존성 예시
func NewA(b *B) *A { return &A{b: b} }
func NewB(a *A) *B { return &B{a: a} }

// 런타임에 에러 발생:
// "cycle detected in dependency graph"
```

### 누락된 의존성
필요한 Provider가 없으면 에러가 발생합니다.

```go
// Provider 등록 없이 사용하려 하면:
// "missing type: *Database"
```

## 9. FX vs Dig 비교

### Dig (저수준)
```go
container := dig.New()
container.Provide(NewDatabase)
container.Provide(NewService)
container.Invoke(func(service *Service) {
    service.Start()
})
```

### FX (고수준)
```go
fx.New(
    fx.Provide(NewDatabase),
    fx.Provide(NewService),
    fx.Invoke(func(service *Service) {
        service.Start()
    }),
).Run()  // 라이프사이클 자동 관리
```

## 10. 디버깅과 시각화

### 의존성 그래프 시각화
```go
// Dig Container의 의존성 그래프를 DOT 형식으로 출력
container.Visualize(os.Stdout, dig.VisualizeError)
```

### 본 프로젝트 의존성 구조
```
main
├── zap.Logger
├── validator.Validator
├── middleware.LoggerMiddleware ← zap.Logger
├── middleware.ErrorMiddleware ← zap.Logger
├── server.EchoServer ← Lifecycle, Logger, Middlewares, []RouteRegistrar
├── jwt.Service("access") ← zap.Logger
├── jwt.Service("refresh") ← zap.Logger
├── jwt.Service("admin") ← zap.Logger
├── user.Routes ← Handler, UserMiddleware, AdminMiddleware
├── coupon.Routes ← Handler, UserMiddleware, AdminMiddleware
└── auth.Routes ← Handler
```

이러한 Dig의 개념들을 이해하면 FX를 더 효과적으로 활용할 수 있습니다.