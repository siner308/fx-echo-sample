# 보안 가이드

이 문서는 fx-echo-sample 프로젝트의 보안 구현 및 모범 사례를 설명합니다.

## 보안 개요

### 보안 원칙
1. **방어 심층화** (Defense in Depth)
2. **최소 권한 원칙** (Principle of Least Privilege)
3. **기본적으로 거부** (Deny by Default)
4. **입력 검증** (Input Validation)
5. **보안 설계** (Security by Design)

### 보안 레이어
```
┌─────────────────────────────────────────────────────────────┐
│                    Network Security                        │
│                   (HTTPS, Firewall)                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────────────────────────────────────────────┐
│                Application Security                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   JWT Auth  │ │    CORS     │ │  Rate Limit │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────────────────────────────────────────────┐
│                   Data Security                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │  Argon2id   │ │ Environment │ │  Input Val  │           │
│  │  Password   │ │  Variables  │ │  idation    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 패스워드 보안

### Argon2id 해싱
프로젝트는 업계 표준인 Argon2id 알고리즘을 사용하여 패스워드를 안전하게 저장합니다.

#### 구현 세부사항
```go
// pkg/security/password.go
type PasswordConfig struct {
    Memory      uint32 `json:"memory"`      // 메모리 사용량 (KB)
    Iterations  uint32 `json:"iterations"`  // 반복 횟수
    Parallelism uint8  `json:"parallelism"` // 병렬 스레드 수
    SaltLength  uint32 `json:"salt_length"` // 소금 길이
    KeyLength   uint32 `json:"key_length"`  // 출력 키 길이
}

// 기본 설정 (OWASP 권장사항 기준)
func DefaultPasswordConfig() *PasswordConfig {
    return &PasswordConfig{
        Memory:      64 * 1024, // 64MB
        Iterations:  3,         // 3회 반복
        Parallelism: 2,         // 2개 스레드
        SaltLength:  16,        // 16바이트 소금
        KeyLength:   32,        // 32바이트 키
    }
}
```

#### 해싱 과정
```go
func HashPassword(password string, config *PasswordConfig) (string, error) {
    // 1. 설정 검증
    if config == nil {
        config = DefaultPasswordConfig()
    }
    
    // 2. 암호학적으로 안전한 소금 생성
    salt, err := generateRandomBytes(config.SaltLength)
    if err != nil {
        return "", fmt.Errorf("failed to generate salt: %w", err)
    }
    
    // 3. Argon2id 해싱
    hash := argon2.IDKey(
        []byte(password),
        salt,
        config.Iterations,
        config.Memory,
        config.Parallelism,
        config.KeyLength,
    )
    
    // 4. Base64 인코딩 및 메타데이터 포함
    encoded := base64.RawStdEncoding.EncodeToString(hash)
    format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
    
    return fmt.Sprintf(format,
        argon2.Version,
        config.Memory,
        config.Iterations,
        config.Parallelism,
        base64.RawStdEncoding.EncodeToString(salt),
        encoded,
    ), nil
}
```

#### 보안 특징
1. **소금(Salt)**: 각 패스워드마다 고유한 소금 사용
2. **타이밍 공격 방지**: 일정한 시간 소요로 타이밍 공격 방지
3. **메모리 하드**: 메모리 집약적 연산으로 GPU 공격 어려움
4. **설정 가능**: 하드웨어에 맞춰 보안 매개변수 조정 가능

### 패스워드 정책
```go
// 권장 패스워드 정책
const (
    MinPasswordLength = 8
    MaxPasswordLength = 128
)

// 패스워드 복잡성 검증 (향후 구현)
type PasswordPolicy struct {
    RequireUppercase bool
    RequireLowercase bool
    RequireDigits    bool
    RequireSymbols   bool
    MinLength       int
    MaxLength       int
}
```

## JWT 인증 보안

### 토큰 보안 설계
```go
// pkg/jwt/service.go
type Claims struct {
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role,omitempty"`
    jwt.RegisteredClaims
}

// 토큰 생성 시 보안 체크리스트
func (s *service) GenerateToken(userID int, email string, role ...string) (string, error) {
    now := time.Now()
    
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(s.config.ExpiresIn)), // 만료 시간
            IssuedAt:  jwt.NewNumericDate(now),                         // 발급 시간
            NotBefore: jwt.NewNumericDate(now),                         // 유효 시작 시간
            Issuer:    s.config.Issuer,                                 // 발급자
            Subject:   email,                                           // 주체
            ID:        s.config.TokenType,                              // 토큰 타입
        },
    }
    
    // HMAC-SHA256 서명
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.Secret))
}
```

### 토큰 검증 보안
```go
func (s *service) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // 1. 서명 방법 검증
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(s.config.Secret), nil
    })
    
    if err != nil {
        // 2. 만료 토큰 처리
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidClaims
    }
    
    // 3. 토큰 타입 검증
    if claims.ID != s.config.TokenType {
        return nil, ErrInvalidToken
    }
    
    return claims, nil
}
```

### JWT 보안 모범 사례
1. **짧은 만료 시간**: Access Token 1시간, Refresh Token 24시간
2. **강력한 시크릿**: 최소 256비트 랜덤 키
3. **토큰 타입 구분**: Access, Refresh, Admin 토큰 분리
4. **클레임 검증**: 발급자, 만료시간, 토큰 타입 검증

## 환경변수 보안

### 보안 환경변수 관리
```go
// 기본값 없는 필수 환경변수
func getRequiredEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        panic(fmt.Sprintf("Required environment variable %s is not set", key))
    }
    return value
}

// 보안이 중요한 환경변수들
var requiredSecrets = []string{
    "ACCESS_TOKEN_SECRET",
    "REFRESH_TOKEN_SECRET", 
    "ADMIN_TOKEN_SECRET",
    "DATABASE_PASSWORD",
}
```

### 환경변수 보안 원칙
1. **기본값 금지**: 보안 관련 변수는 기본값 제공하지 않음
2. **최소 길이**: JWT 시크릿은 최소 32바이트
3. **로그 제외**: 환경변수는 로그에 출력하지 않음
4. **컨테이너 시크릿**: Docker/Kubernetes secrets 사용

### .env 파일 보안
```bash
# .env.example (공개 가능)
ACCESS_TOKEN_SECRET=your_access_token_secret_here
REFRESH_TOKEN_SECRET=your_refresh_token_secret_here
ADMIN_TOKEN_SECRET=your_admin_token_secret_here

# .env (실제 파일은 .gitignore에 포함)
ACCESS_TOKEN_SECRET=actual_secret_key_here
REFRESH_TOKEN_SECRET=actual_refresh_secret_here
ADMIN_TOKEN_SECRET=actual_admin_secret_here
```

## 입력 검증 및 보안

### Echo Context 보안
```go
// 타입 안전 컨텍스트 키
type UserContextKey string
const UserIDKey UserContextKey = "user_id"

// 안전한 컨텍스트 사용
func setUserContext(c echo.Context, userID int) {
    c.Set(string(UserIDKey), userID)
}

func getUserFromContext(c echo.Context) (int, error) {
    userID, ok := c.Get(string(UserIDKey)).(int)
    if !ok {
        return 0, errors.New("user not found in context")
    }
    return userID, nil
}
```

### 입력 검증 (향후 구현)
```go
// go-playground/validator 사용 예시
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=1,max=120"`
    Password string `json:"password" validate:"required,min=8,max=128"`
}

func validateRequest(req interface{}) error {
    validate := validator.New()
    return validate.Struct(req)
}
```

## 미들웨어 보안

### CORS 설정
```go
// middleware/cors.go
func CORS() echo.MiddlewareFunc {
    return middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins:     []string{"https://yourdomain.com"},  // 특정 도메인만 허용
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
        AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
        AllowCredentials: true,
        MaxAge:           86400, // 24시간
    })
}
```

### 보안 헤더
```go
// 보안 헤더 미들웨어 (향후 구현)
func SecurityHeaders() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // XSS 보호
            c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
            
            // 콘텐츠 타입 스니핑 방지
            c.Response().Header().Set("X-Content-Type-Options", "nosniff")
            
            // 클릭재킹 방지
            c.Response().Header().Set("X-Frame-Options", "DENY")
            
            // HTTPS 강제 (프로덕션)
            c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            
            return next(c)
        }
    }
}
```

### Rate Limiting (향후 구현)
```go
// 요청 제한 미들웨어
func RateLimit() echo.MiddlewareFunc {
    return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
        Skipper: middleware.DefaultSkipper,
        Store: middleware.NewRateLimiterMemoryStoreWithConfig(
            middleware.RateLimiterMemoryStoreConfig{
                Rate:      10,               // 초당 10회
                Burst:     20,               // 버스트 20회
                ExpiresIn: time.Minute * 3,  // 3분 후 만료
            },
        ),
        IdentifierExtractor: func(ctx echo.Context) (string, error) {
            id := ctx.RealIP()
            return id, nil
        },
        ErrorHandler: func(context echo.Context, err error) error {
            return context.JSON(http.StatusForbidden, map[string]string{
                "message": "rate limit exceeded",
            })
        },
        DenyHandler: func(context echo.Context, identifier string, err error) error {
            return context.JSON(http.StatusTooManyRequests, map[string]string{
                "message": "too many requests",
            })
        },
    })
}
```

## 로깅 보안

### 보안 로깅 원칙
```go
// 로그에서 민감한 정보 제외
func (s *service) CreateUser(req CreateUserRequest) (*entity.User, error) {
    // 좋은 예: 이메일만 로그
    s.logger.Info("Creating user", zap.String("email", req.Email))
    
    // 나쁜 예: 패스워드 로그 (절대 금지!)
    // s.logger.Info("Creating user", zap.String("password", req.Password))
    
    hashedPassword, err := security.HashPassword(req.Password, nil)
    if err != nil {
        // 에러는 로그하되 패스워드는 제외
        s.logger.Error("Failed to hash password", zap.Error(err))
        return nil, errors.New("failed to process password")
    }
    
    // ...
}
```

### 보안 이벤트 로깅
```go
// 보안 관련 이벤트 로깅
func (s *authService) Login(email, password string) (*LoginResponse, error) {
    user, err := s.passwordVerifier.VerifyUserPassword(email, password)
    if err != nil {
        // 로그인 실패 로깅 (브루트포스 공격 탐지용)
        s.logger.Warn("Login failed", 
            zap.String("email", email),
            zap.String("ip", ""),  // 실제로는 IP 주소 포함
            zap.Error(err))
        return nil, ErrInvalidCredentials
    }
    
    // 성공적인 로그인 로깅
    s.logger.Info("User logged in successfully",
        zap.Int("user_id", user.ID),
        zap.String("email", user.Email))
    
    // ...
}
```

## 데이터베이스 보안 (향후 구현)

### SQL 인젝션 방지
```go
// 좋은 예: Prepared Statement 사용
func (r *repository) GetUserByEmail(email string) (*entity.User, error) {
    query := "SELECT id, name, email, password FROM users WHERE email = ?"
    row := r.db.QueryRow(query, email)
    // ...
}

// 나쁜 예: 문자열 연결 (SQL 인젝션 취약)
// query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
```

### 데이터베이스 연결 보안
```go
// 데이터베이스 연결 시 보안 설정
func connectDB() (*sql.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=true",
        getRequiredEnv("DB_USER"),
        getRequiredEnv("DB_PASSWORD"),
        getRequiredEnv("DB_HOST"),
        getRequiredEnv("DB_PORT"),
        getRequiredEnv("DB_NAME"))
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    
    // 연결 풀 설정
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}
```

## 보안 테스트

### 패스워드 보안 테스트
```go
func TestPasswordSecurity(t *testing.T) {
    password := "testPassword123"
    
    // 1. 해싱 테스트
    hash1, err := HashPassword(password, nil)
    assert.NoError(t, err)
    
    hash2, err := HashPassword(password, nil)
    assert.NoError(t, err)
    
    // 동일한 패스워드라도 다른 해시 생성 (소금 때문)
    assert.NotEqual(t, hash1, hash2)
    
    // 2. 검증 테스트
    isValid, err := VerifyPassword(password, hash1)
    assert.NoError(t, err)
    assert.True(t, isValid)
    
    // 3. 잘못된 패스워드 테스트
    isValid, err = VerifyPassword("wrongPassword", hash1)
    assert.NoError(t, err)
    assert.False(t, isValid)
}
```

### 타이밍 공격 테스트
```go
func TestTimingAttackPrevention(t *testing.T) {
    password := "testPassword123"
    hash, _ := HashPassword(password, nil)
    
    // 올바른 패스워드와 잘못된 패스워드의 처리 시간 측정
    correctTimes := make([]time.Duration, 10)
    incorrectTimes := make([]time.Duration, 10)
    
    for i := 0; i < 10; i++ {
        start := time.Now()
        VerifyPassword(password, hash)
        correctTimes[i] = time.Since(start)
        
        start = time.Now()
        VerifyPassword("wrongPassword", hash)
        incorrectTimes[i] = time.Since(start)
    }
    
    // 시간 차이가 크지 않아야 함 (타이밍 공격 방지)
    avgCorrect := averageDuration(correctTimes)
    avgIncorrect := averageDuration(incorrectTimes)
    
    timeDiff := time.Duration(math.Abs(float64(avgCorrect - avgIncorrect)))
    assert.Less(t, timeDiff, time.Millisecond*10, "Timing difference too large")
}
```

## 보안 체크리스트

### 개발 시 보안 체크리스트
- [ ] 패스워드는 Argon2id로 해싱
- [ ] JWT 시크릿은 환경변수로 관리
- [ ] 환경변수에 기본값 없음
- [ ] 민감한 정보 로그 제외
- [ ] 입력값 검증 및 이스케이프
- [ ] HTTPS 사용 (프로덕션)
- [ ] CORS 적절히 설정
- [ ] 에러 메시지에 민감한 정보 포함 금지

### 배포 시 보안 체크리스트
- [ ] 모든 환경변수 설정 확인
- [ ] HTTPS 인증서 설정
- [ ] 방화벽 규칙 설정
- [ ] 로그 모니터링 설정
- [ ] 정기적인 보안 업데이트
- [ ] 백업 암호화

### 보안 모니터링
- [ ] 로그인 실패 횟수 모니터링
- [ ] 비정상적인 API 호출 패턴 탐지
- [ ] 토큰 남용 탐지
- [ ] 시스템 리소스 사용량 모니터링

## 보안 리소스

### 참고 문서
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [JWT Security Best Practices](https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/)
- [Argon2 RFC](https://datatracker.ietf.org/doc/html/rfc9106)
- [Go Security Guide](https://github.com/securego/gosec)

### 보안 도구
- `gosec`: Go 코드 보안 분석
- `safety`: 의존성 보안 검사
- `trivy`: 컨테이너 보안 스캔