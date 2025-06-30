# 테스트 가이드

이 문서는 fx-echo-sample 프로젝트의 테스트 전략과 실행 방법을 설명합니다.

## 테스트 전략

### 테스트 피라미드
```
        ┌─────────────────┐
        │   E2E Tests     │  ← 소수의 전체 플로우 테스트
        └─────────────────┘
       ┌───────────────────┐
       │ Integration Tests │   ← 모듈 간 상호작용 테스트
       └───────────────────┘
     ┌─────────────────────────┐
     │     Unit Tests          │  ← 다수의 개별 함수/메서드 테스트
     └─────────────────────────┘
```

### 테스트 종류

1. **Unit Tests**: 개별 함수/메서드의 동작 검증
2. **Integration Tests**: 모듈 간 상호작용 검증
3. **End-to-End Tests**: 전체 API 플로우 검증
4. **Benchmark Tests**: 성능 측정

## 테스트 실행

### 전체 테스트 실행
```bash
# 모든 테스트 실행
go test ./...

# 상세 출력과 함께 실행
go test -v ./...

# 커버리지 포함 실행
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 모듈별 테스트 실행
```bash
# 보안 패키지 테스트
go test -v ./pkg/security/

# 사용자 서비스 테스트
go test -v ./modules/user/

# 인증 서비스 테스트
go test -v ./modules/auth/user/
```

### 벤치마크 테스트 실행
```bash
# 성능 벤치마크 실행
go test -bench=. ./pkg/security/

# 메모리 프로파일링 포함
go test -bench=. -memprofile=mem.prof ./pkg/security/
```

## 테스트 구현

### 1. Security Package 테스트

#### 패스워드 해싱 테스트
```go
func TestHashPassword(t *testing.T) {
    password := "testPassword123"
    
    hash, err := HashPassword(password, nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.NotEqual(t, password, hash)
    
    // 동일한 패스워드로 다시 해싱 시 다른 결과
    hash2, err := HashPassword(password, nil)
    assert.NoError(t, err)
    assert.NotEqual(t, hash, hash2)
}
```

#### 패스워드 검증 테스트
```go
func TestVerifyPassword(t *testing.T) {
    password := "testPassword123"
    hash, _ := HashPassword(password, nil)
    
    // 올바른 패스워드 검증
    isValid, err := VerifyPassword(password, hash)
    assert.NoError(t, err)
    assert.True(t, isValid)
    
    // 잘못된 패스워드 검증
    isValid, err = VerifyPassword("wrongPassword", hash)
    assert.NoError(t, err)
    assert.False(t, isValid)
}
```

#### 벤치마크 테스트
```go
func BenchmarkHashPassword(b *testing.B) {
    password := "benchmarkPassword123"
    config := DefaultPasswordConfig()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := HashPassword(password, config)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 2. User Service 테스트

#### Mock Repository 정의
```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *entity.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*entity.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.User), args.Error(1)
}
```

#### 사용자 생성 테스트
```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name        string
        request     CreateUserRequest
        setupMock   func(*MockUserRepository)
        wantErr     bool
        wantErrType error
    }{
        {
            name: "successful user creation",
            request: CreateUserRequest{
                Name:     "John Doe",
                Email:    "john@example.com",
                Age:      30,
                Password: "password123",
            },
            setupMock: func(m *MockUserRepository) {
                m.On("Create", mock.AnythingOfType("*entity.User")).Return(nil)
            },
            wantErr: false,
        },
        // ... 더 많은 테스트 케이스
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockUserRepository)
            tt.setupMock(mockRepo)
            
            service := setupUserService(mockRepo)
            user, err := service.CreateUser(tt.request)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                // 패스워드가 해싱되었는지 확인
                assert.NotEqual(t, tt.request.Password, user.Password)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

#### 패스워드 검증 테스트
```go
func TestVerifyUserPassword(t *testing.T) {
    // 실제 해싱된 패스워드로 테스트 사용자 생성
    plainPassword := "testPassword123"
    hashedPassword, err := security.HashPassword(plainPassword, nil)
    assert.NoError(t, err)
    
    testUser := &entity.User{
        ID:       1,
        Email:    "test@example.com",
        Password: hashedPassword,
    }
    
    mockRepo := new(MockUserRepository)
    mockRepo.On("GetByEmail", "test@example.com").Return(testUser, nil)
    
    service := setupUserService(mockRepo)
    user, err := service.VerifyUserPassword("test@example.com", plainPassword)
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, testUser.ID, user.ID)
}
```

### 3. Auth Service 테스트

#### Mock JWT Service 정의
```go
type MockJWTService struct {
    mock.Mock
}

func (m *MockJWTService) GenerateToken(userID int, email string, roles ...string) (string, error) {
    var mockArgs []interface{}
    mockArgs = append(mockArgs, userID, email)
    for _, role := range roles {
        mockArgs = append(mockArgs, role)
    }
    args := m.Called(mockArgs...)
    return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwt.Claims, error) {
    args := m.Called(tokenString)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*jwt.Claims), args.Error(1)
}
```

#### 로그인 테스트
```go
func TestLogin(t *testing.T) {
    testUser := UserInfo{
        ID:    1,
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    tests := []struct {
        name      string
        email     string
        password  string
        setupMock func(*MockJWTService, *MockJWTService, *MockPasswordVerifier)
        wantErr   bool
    }{
        {
            name:     "successful login",
            email:    "test@example.com",
            password: "password123",
            setupMock: func(access, refresh *MockJWTService, verifier *MockPasswordVerifier) {
                verifier.On("VerifyUserPassword", "test@example.com", "password123").Return(testUser, nil)
                access.On("GenerateToken", 1, "test@example.com").Return("access_token", nil)
                refresh.On("GenerateToken", 1, "test@example.com").Return("refresh_token", nil)
                access.On("GetExpirationTime").Return(time.Hour)
            },
            wantErr: false,
        },
        // ... 더 많은 테스트 케이스
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 테스트 실행 로직
        })
    }
}
```

#### 토큰 갱신 테스트
```go
func TestRefreshToken(t *testing.T) {
    validClaims := &jwt.Claims{
        UserID: 1,
        Email:  "test@example.com",
        Role:   "user",
    }
    
    mockAccess := new(MockJWTService)
    mockRefresh := new(MockJWTService)
    
    mockRefresh.On("ValidateToken", "valid_refresh_token").Return(validClaims, nil)
    mockAccess.On("GenerateToken", 1, "test@example.com", "user").Return("new_access_token", nil)
    mockAccess.On("GetExpirationTime").Return(time.Hour)
    
    service := setupAuthService(mockAccess, mockRefresh, nil)
    response, err := service.RefreshToken("valid_refresh_token")
    
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, "new_access_token", response.AccessToken)
}
```

### 4. 동시성 테스트

#### 동시 로그인 테스트
```go
func TestConcurrentLogin(t *testing.T) {
    // ... setup code
    
    const numGoroutines = 10
    done := make(chan bool, numGoroutines)
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func() {
            _, err := service.Login("test@example.com", "password123")
            if err != nil {
                errors <- err
            }
            done <- true
        }()
    }
    
    // 모든 고루틴 완료 대기
    for i := 0; i < numGoroutines; i++ {
        <-done
    }
    
    // 에러 검사
    close(errors)
    for err := range errors {
        t.Errorf("Concurrent login failed: %v", err)
    }
}
```

## 테스트 도구 및 라이브러리

### 사용 중인 테스트 라이브러리
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"    // 어설션
    "github.com/stretchr/testify/mock"      // 모킹
    "go.uber.org/zap"                       // 로깅 (테스트용 NOP)
)
```

### 어설션 예제
```go
// 기본 어설션
assert.NoError(t, err)
assert.Error(t, err)
assert.Equal(t, expected, actual)
assert.NotEqual(t, unexpected, actual)
assert.True(t, condition)
assert.False(t, condition)
assert.Nil(t, object)
assert.NotNil(t, object)

// 컬렉션 어설션
assert.Len(t, slice, expectedLength)
assert.Contains(t, slice, item)
assert.Empty(t, slice)

// 에러 타입 검사
assert.ErrorIs(t, err, expectedErrorType)
```

### Mock 사용 패턴
```go
// Mock 설정
mockRepo.On("MethodName", arg1, arg2).Return(result, nil)
mockRepo.On("MethodName", mock.AnythingOfType("*Type")).Return(result, nil)

// Mock 실행 후 검증
mockRepo.AssertExpectations(t)
mockRepo.AssertCalled(t, "MethodName", arg1, arg2)
mockRepo.AssertNumberOfCalls(t, "MethodName", 1)
```

## 테스트 커버리지

### 현재 커버리지 현황
- **Security Package**: 100% (모든 공개 함수)
- **User Service**: 95% (주요 비즈니스 로직)
- **Auth Service**: 90% (인증 플로우)
- **Integration Tests**: 80% (API 엔드포인트)

### 커버리지 목표
- Unit Tests: 90% 이상
- Integration Tests: 80% 이상
- Critical Paths: 100%

## 테스트 베스트 프랙티스

### 1. 테스트 네이밍
```go
// 좋은 예
func TestCreateUser_WithValidData_ReturnsUser(t *testing.T) {}
func TestLogin_WithInvalidCredentials_ReturnsError(t *testing.T) {}

// 나쁜 예
func TestCreateUser(t *testing.T) {}
func TestLogin(t *testing.T) {}
```

### 2. 테스트 구조 (AAA 패턴)
```go
func TestSomething(t *testing.T) {
    // Arrange (준비)
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    
    // Act (실행)
    result, err := service.DoSomething()
    
    // Assert (검증)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 3. 테스트 격리
- 각 테스트는 독립적이어야 함
- 공유 상태 사용 금지
- 테스트 순서에 의존 금지

### 4. Mock vs Real Objects
```go
// Unit Test: Mock 사용
func TestService_WithMockRepo() {
    mockRepo := new(MockRepository)
    // ...
}

// Integration Test: Real Objects 사용
func TestService_WithRealRepo() {
    realRepo := NewInMemoryRepository()
    // ...
}
```

## CI/CD 통합

### GitHub Actions 예제
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go test -v -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage
      uses: actions/upload-artifact@v2
      with:
        name: coverage
        path: coverage.html
```

## 문제 해결

### 일반적인 테스트 이슈

1. **Mock 설정 오류**
   ```bash
   panic: mock: Unexpected Method Call
   ```
   → Mock.On() 호출에서 매개변수 타입/값 확인

2. **타이밍 이슈**
   ```bash
   test timeout after 10s
   ```
   → 동시성 테스트에서 데드락 또는 무한 대기

3. **환경변수 이슈**
   ```bash
   panic: Required environment variable not set
   ```
   → 테스트용 환경변수 설정 또는 Mock 사용

### 디버깅 팁
```go
// 테스트 실패 시 추가 정보 출력
t.Logf("Expected: %v, Got: %v", expected, actual)

// 임시 디버깅용 출력
fmt.Printf("Debug: %+v\n", variable)

// 테스트 스킵
t.Skip("Skipping this test for now")
```

## 성능 테스트

### 벤치마크 결과 예시
```
BenchmarkHashPassword-8         20    54.2ms/op    32768 B/op    4 allocs/op
BenchmarkVerifyPassword-8       18    56.1ms/op    32768 B/op    4 allocs/op
```

### 성능 목표
- 패스워드 해싱: < 100ms
- API 응답 시간: < 100ms
- 동시 사용자: 100명 이상