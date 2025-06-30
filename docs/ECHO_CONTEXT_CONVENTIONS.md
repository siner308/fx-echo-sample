# Echo Context 사용 컨벤션

## 📋 목적
Echo Context(`echo.Context`)의 올바른 사용법과 클린 아키텍처를 유지하기 위한 컨벤션을 정의합니다.

## 🎯 핵심 원칙

### 1. **레이어 경계 준수**
```go
// ✅ 올바른 사용: HTTP 레이어에서만 사용
func (h *Handler) GetUser(echoCtx echo.Context) error {
    userID := getUserIDFromPath(echoCtx)  // OK: 파라미터 추출
    
    user, err := h.service.GetUser(userID)  // OK: 비즈니스 로직에 전달하지 않음
    if err != nil {
        return echoCtx.JSON(500, ErrorResponse{Error: err.Error()})
    }
    
    return echoCtx.JSON(200, user)  // OK: 응답 생성
}

// ❌ 잘못된 사용: 서비스 레이어에 전달
func (h *Handler) GetUser(echoCtx echo.Context) error {
    return h.service.GetUserWithContext(echoCtx)  // 금지!
}
```

### 2. **변수명 규칙**
```go
// ✅ 명확한 변수명
func middlewareFunc(echoCtx echo.Context) error
func handlerFunc(echoCtx echo.Context) error

// ✅ Go context와 구분
func someFunction(echoCtx echo.Context) error {
    reqCtx := echoCtx.Request().Context()  // Go context
    // echoCtx = Echo 관련, reqCtx = Go context 작업
}

// ❌ 혼동을 야기하는 변수명
func middlewareFunc(c echo.Context) error     // 금지
func middlewareFunc(ctx echo.Context) error   // 금지
```

## 🏗️ 아키텍처 레이어별 사용 규칙

### **HTTP Layer (Handlers, Middleware)**
```go
// ✅ 허용되는 작업들
func (h *Handler) CreateUser(echoCtx echo.Context) error {
    // 1. 요청 데이터 바인딩
    var req CreateUserRequest
    if err := echoCtx.Bind(&req); err != nil {
        return badRequestError(echoCtx, "Invalid request")
    }
    
    // 2. 컨텍스트에서 인증 정보 추출
    userID, _ := user.GetUserID(echoCtx)
    
    // 3. 비즈니스 로직 호출 (echo context 전달하지 않음)
    result, err := h.service.CreateUser(req, userID)
    
    // 4. 응답 생성
    return echoCtx.JSON(201, result)
}

// ✅ 미들웨어에서의 컨텍스트 데이터 설정
func (m *Middleware) AuthMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(echoCtx echo.Context) error {
            // 인증 처리 후 컨텍스트에 설정
            echoCtx.Set(string(UserIDKey), claims.UserID)
            return next(echoCtx)
        }
    }
}
```

### **Service Layer (Business Logic)**
```go
// ✅ 올바른 서비스 인터페이스
type UserService interface {
    CreateUser(req CreateUserRequest, currentUserID int) (*User, error)
    GetUser(id int) (*User, error)
    UpdateUser(id int, req UpdateUserRequest) (*User, error)
}

// ❌ Echo Context를 받는 서비스 (금지)
type UserService interface {
    CreateUser(echoCtx echo.Context) (*User, error)  // 금지!
    GetUser(echoCtx echo.Context) (*User, error)     // 금지!
}
```

### **Repository Layer (Data Access)**
```go
// ✅ 데이터베이스 컨텍스트는 Go context 사용
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int) (*User, error)
    Update(ctx context.Context, user *User) error
}

// ✅ 서비스에서 Go context 전달
func (s *service) CreateUser(req CreateUserRequest, currentUserID int) (*User, error) {
    ctx := context.Background()  // 또는 timeout context
    return s.repo.Create(ctx, user)
}
```

## 🔧 Context 데이터 관리

### **Type-safe Context Keys**
```go
// ✅ 각 모듈별로 Context Key 타입 정의
type UserContextKey string
type AdminContextKey string

const (
    UserIDKey    UserContextKey = "user_id"
    UserEmailKey UserContextKey = "user_email"
)

// ✅ Helper 함수 제공
func GetUserID(echoCtx echo.Context) (int, bool) {
    userID, ok := echoCtx.Get(string(UserIDKey)).(int)
    return userID, ok
}
```

### **Context 데이터 범위**
```go
// ✅ 인증/인가 정보만 저장
echoCtx.Set(string(UserIDKey), claims.UserID)
echoCtx.Set(string(UserRoleKey), claims.Role)

// ❌ 비즈니스 데이터 저장 금지
echoCtx.Set("user_profile", userProfile)  // 금지!
echoCtx.Set("cached_data", someData)      // 금지!
```

## 🚫 금지사항

### **1. Service Layer에 Echo Context 전달**
```go
// ❌ 금지
func (h *Handler) GetUser(echoCtx echo.Context) error {
    return h.service.ProcessUser(echoCtx)  // 레이어 경계 위반
}

// ✅ 올바른 방법
func (h *Handler) GetUser(echoCtx echo.Context) error {
    userID := getUserIDFromPath(echoCtx)
    user, err := h.service.GetUser(userID)  // 필요한 데이터만 전달
    return respondWithUser(echoCtx, user, err)
}
```

### **2. Echo Context를 다른 Goroutine에 전달**
```go
// ❌ 금지 - Race condition 위험
func (h *Handler) AsyncProcess(echoCtx echo.Context) error {
    go func() {
        result := processData()
        echoCtx.JSON(200, result)  // 위험!
    }()
    return nil
}

// ✅ 올바른 방법
func (h *Handler) AsyncProcess(echoCtx echo.Context) error {
    reqCtx := echoCtx.Request().Context()
    go func(ctx context.Context) {
        result := processDataWithContext(ctx)
        // 결과는 채널이나 다른 방법으로 전달
    }(reqCtx)
    return echoCtx.JSON(202, map[string]string{"status": "processing"})
}
```

### **3. 전역 변수에 Echo Context 저장**
```go
// ❌ 절대 금지
var globalEchoCtx echo.Context  // 위험!

func (h *Handler) SetGlobalContext(echoCtx echo.Context) error {
    globalEchoCtx = echoCtx  // 절대 금지!
    return nil
}
```

## 📝 Helper 함수 패턴

### **Request 데이터 추출**
```go
// ✅ 재사용 가능한 헬퍼 함수들
func getUserIDFromPath(echoCtx echo.Context) (int, error) {
    idStr := echoCtx.Param("id")
    return strconv.Atoi(idStr)
}

func bindAndValidate[T any](echoCtx echo.Context, req *T, validator *validator.Validator) error {
    if err := echoCtx.Bind(req); err != nil {
        return echo.NewHTTPError(400, "Invalid request format")
    }
    if err := validator.Validate(req); err != nil {
        return echo.NewHTTPError(400, "Validation failed")
    }
    return nil
}
```

### **Response 생성**
```go
// ✅ 표준화된 응답 헬퍼
func respondWithError(echoCtx echo.Context, err error) error {
    if errors.Is(err, ErrNotFound) {
        return echoCtx.JSON(404, ErrorResponse{Error: "Resource not found"})
    }
    return echoCtx.JSON(500, ErrorResponse{Error: "Internal server error"})
}

func respondWithData(echoCtx echo.Context, data interface{}) error {
    return echoCtx.JSON(200, data)
}
```

## ✅ 체크리스트

### **코드 리뷰시 확인사항**
- [ ] Echo Context가 HTTP 레이어를 벗어나지 않는가?
- [ ] 변수명이 `echoCtx`로 명확히 명명되었는가?
- [ ] Context 데이터는 type-safe key를 사용하는가?
- [ ] 서비스 레이어에 필요한 데이터만 전달하는가?
- [ ] Echo Context를 다른 goroutine에 전달하지 않는가?

### **리팩토링 가이드**
1. Echo Context를 받는 서비스 메서드 찾기
2. 필요한 데이터만 매개변수로 추출
3. Echo Context 의존성 제거
4. 테스트 가능성 향상 확인

## 🎯 결론

**Echo Context는 HTTP 레이어의 도구입니다.**
- 요청/응답 처리에만 사용
- 레이어 경계를 넘지 않음
- 명확한 변수명 사용
- Type-safe 데이터 접근

이 컨벤션을 따르면 **테스트 가능하고 유지보수하기 쉬운 클린 아키텍처**를 구축할 수 있습니다.