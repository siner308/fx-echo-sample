# FX 트러블슈팅 가이드

본 프로젝트 개발 중 실제로 발생한 문제들과 해결 방법, 그리고 일반적인 FX 문제들을 정리합니다.

## 1. 순환 의존성 (Circular Dependency)

### 🚨 문제: Auth Service ↔ User Service 순환 참조

**발생 상황:**
```go
// modules/auth/user/service.go
func NewService(userService user.Service) Service {
    return &service{userService: userService}
}

// modules/user/service.go  
func NewService(authService auth.Service) Service {
    return &service{authService: authService}
}
```

**에러 메시지:**
```
cycle detected in dependency graph
```

**해결 방법:**
```go
// ✅ Repository 직접 사용으로 순환 의존성 제거
// modules/auth/user/service.go
type ServiceParam struct {
    fx.In
    Repository    repository.Repository  // Service 대신 Repository 직접 사용
    AccessToken   jwt.Service `name:"access"`
    RefreshToken  jwt.Service `name:"refresh"`
    Logger        *zap.Logger
}
```

### 일반적인 순환 의존성 해결 패턴

#### 1. 인터페이스 분리
```go
// 공통 인터페이스 정의
type UserRepository interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}

// Auth Service는 Repository만 의존
func NewAuthService(repo UserRepository) AuthService

// User Service는 Repository만 의존  
func NewUserService(repo UserRepository) UserService
```

#### 2. Event 기반 분리
```go
// 이벤트 버스를 통한 느슨한 결합
type EventBus interface {
    Publish(event Event)
    Subscribe(eventType string, handler Handler)
}

func NewAuthService(eventBus EventBus) AuthService
func NewUserService(eventBus EventBus) UserService
```

## 2. Missing Provider 에러

### 🚨 문제: 필요한 Provider 누락

**에러 메시지:**
```
missing type: *jwt.Service (name:"access")
```

**원인 분석:**
```go
// modules/auth/user/service.go에서 요구
type ServiceParam struct {
    fx.In
    AccessToken jwt.Service `name:"access"`  // 이 Provider가 없음
}
```

**해결 방법:**
```go
// ✅ 누락된 Provider 추가
// main.go에 jwt.Providers 추가
func main() {
    fx.New(
        fx.Provide(
            zap.NewProduction,
            // ...
        ),
        jwt.Providers,  // 이 모듈이 누락되어 있었음
        auth.Module,
        user.Module,
        coupon.Module,
    ).Run()
}
```

### Provider 누락 디버깅 방법

#### 1. 의존성 그래프 출력
```go
// 개발용 의존성 확인 함수
func checkDependencies() {
    app := fx.New(
        // 모든 모듈 포함
        jwt.Providers,
        auth.Module,
        user.Module,
        fx.Invoke(func() {
            log.Println("All dependencies resolved")
        }),
    )
    
    if err := app.Err(); err != nil {
        log.Fatal("Dependency resolution failed:", err)
    }
}
```

#### 2. 단계별 모듈 추가
```go
// 하나씩 추가하며 어느 시점에서 실패하는지 확인
fx.New(
    fx.Provide(zap.NewProduction),
    // jwt.Providers,  // 주석 처리하고 테스트
    auth.Module,
).Run()
```

## 3. 환경변수 설정 문제

### 🚨 문제: 환경변수 기본값으로 인한 보안 위험

**위험한 코드:**
```go
// ❌ 기본값이 보안 위험을 초래
func NewJWTService() Service {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "default-secret"  // 프로덕션에서 위험!
    }
    return &service{secret: secret}
}
```

**해결 방법:**
```go
// ✅ 필수 환경변수 강제
func getRequiredEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        panic("Required environment variable not set: " + key)
    }
    return value
}

func NewJWTService() Service {
    return &service{
        secret: getRequiredEnv("JWT_SECRET"),  // 필수값 검증
    }
}
```

### 환경변수 설정 체크리스트

#### 1. .env.example 파일 제공
```bash
# JWT 설정
ACCESS_TOKEN_SECRET=your-access-token-secret-here
REFRESH_TOKEN_SECRET=your-refresh-token-secret-here
ADMIN_TOKEN_SECRET=your-admin-token-secret-here
ACCESS_TOKEN_EXPIRES=24h
REFRESH_TOKEN_EXPIRES=168h
ADMIN_TOKEN_EXPIRES=8h
JWT_ISSUER=fxserver

# Keycloak 설정 (선택사항)
KEYCLOAK_BASE_URL=http://localhost:8080
KEYCLOAK_REALM=your-realm
KEYCLOAK_CLIENT_ID=your-client-id
KEYCLOAK_CLIENT_SECRET=your-client-secret
```

#### 2. 환경변수 검증 함수
```go
func validateEnvironment() {
    required := []string{
        "ACCESS_TOKEN_SECRET",
        "REFRESH_TOKEN_SECRET", 
        "ADMIN_TOKEN_SECRET",
        "JWT_ISSUER",
    }
    
    for _, key := range required {
        if os.Getenv(key) == "" {
            log.Fatalf("Required environment variable %s is not set", key)
        }
    }
}
```

## 4. GOPATH 설정 문제

### 🚨 문제: IDE 자동완성 실패

**증상:**
- `os.Getenv` 자동완성 후 `os.` 텍스트 남음
- Import 경로 인식 실패
- 빌드는 되지만 IDE에서 에러 표시

**해결 방법:**
```bash
# GOPATH 조정 (사용자 피드백에 의한 해결)
export GOPATH=/path/to/your/workspace
# 또는 IDE 설정에서 Go 모듈 경로 재설정
```

### Go 모듈 관련 문제 해결

#### 1. 모듈 초기화 확인
```bash
go mod init fxserver
go mod tidy
```

#### 2. IDE 설정 확인
- VSCode: Go extension 설정 확인
- GoLand: GOROOT, GOPATH 설정 확인
- go.mod 파일이 프로젝트 루트에 있는지 확인

## 5. Named Dependency 문제

### 🚨 문제: 잘못된 Name 태그

**에러 상황:**
```go
// Provider에서는 "access"로 제공
fx.ResultTags(`name:"access"`)

// Consumer에서는 "access_token"으로 요구  
AccessToken jwt.Service `name:"access_token"`  // 이름 불일치!
```

**해결 방법:**
```go
// ✅ 이름 통일
// Provider
fx.ResultTags(`name:"access"`)

// Consumer  
AccessToken jwt.Service `name:"access"`  // 동일한 이름 사용
```

### Named Dependency 모범 사례

#### 1. 상수로 이름 관리
```go
// pkg/jwt/constants.go
const (
    AccessTokenName  = "access"
    RefreshTokenName = "refresh"
    AdminTokenName   = "admin"
)

// Provider
fx.ResultTags(`name:"` + AccessTokenName + `"`)

// Consumer
AccessToken jwt.Service `name:"access"`
```

#### 2. 이름 규칙 문서화
```go
// JWT Token Name Conventions:
// - "access": 사용자 API 접근용 토큰
// - "refresh": Access Token 갱신용 토큰  
// - "admin": 관리자 API 접근용 토큰
```

## 6. Group Collection 문제

### 🚨 문제: 빈 Group 슬라이스

**증상:**
```go
// RouteRegistrars가 항상 빈 슬라이스
RouteRegistrars []router.RouteRegistrar `group:"routes"`
```

**원인 분석:**
```go
// 모듈에서 group 태그 누락
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),
    // fx.ResultTags(`group:"routes"`),  // 이 줄이 누락됨
)
```

**해결 방법:**
```go
// ✅ 모든 RouteRegistrar에 group 태그 추가
fx.Annotate(
    NewRoutes,
    fx.As(new(router.RouteRegistrar)),
    fx.ResultTags(`group:"routes"`),  // group 태그 추가
)
```

## 7. Interface As 문제

### 🚨 문제: Interface 타입 불일치

**에러 메시지:**
```
*Routes does not implement router.RouteRegistrar
```

**원인 확인:**
```go
// 인터페이스 정의
type RouteRegistrar interface {
    RegisterRoutes(e *echo.Echo)
}

// 구현체 확인
func (r *Routes) RegisterRoutes(e *echo.Echo) {  // 메서드 시그니처 일치 확인
    // 구현
}
```

**일반적인 실수:**
```go
// ❌ 메서드 시그니처 불일치
func (r *Routes) RegisterRoutes(engine *gin.Engine) {  // 잘못된 타입
    // gin.Engine 대신 echo.Echo를 사용해야 함
}
```

## 8. 디버깅 도구와 팁

### 1. 의존성 그래프 시각화
```go
import "go.uber.org/fx/fxevent"

fx.New(
    // 모든 모듈
    fx.WithLogger(func() fxevent.Logger {
        return &fxevent.ConsoleLogger{W: os.Stdout}
    }),
    fx.Invoke(func() {
        log.Println("App started successfully")
    }),
)
```

### 2. Provider 등록 확인
```go
// 개발용 Provider 목록 출력
fx.New(
    fx.Provide(
        func() string {
            log.Println("String provider registered")
            return "test"
        },
    ),
    fx.Invoke(func(s string) {
        log.Printf("Received string: %s", s)
    }),
)
```

### 3. 조건부 디버깅
```go
func main() {
    options := []fx.Option{
        fx.Provide(zap.NewProduction),
        auth.Module,
        user.Module,
    }
    
    if os.Getenv("DEBUG") == "true" {
        options = append(options, fx.WithLogger(func() fxevent.Logger {
            return &fxevent.ConsoleLogger{W: os.Stdout}
        }))
    }
    
    fx.New(options...).Run()
}
```

## 9. 성능 문제 해결

### 1. 불필요한 Provider 제거
```go
// ❌ 사용하지 않는 Provider들
fx.Provide(
    NewHeavyService1,  // 사용하지 않는 무거운 서비스
    NewHeavyService2,  // 사용하지 않는 무거운 서비스
)
```

### 2. 지연 초기화 적용
```go
// ✅ 선택적 초기화
func NewOptionalService() *OptionalService {
    if os.Getenv("ENABLE_OPTIONAL_FEATURE") != "true" {
        return nil
    }
    return &OptionalService{}
}
```

이러한 문제들과 해결책을 참고하여 FX 개발 시 발생할 수 있는 문제들을 미리 예방하고 빠르게 해결할 수 있습니다.