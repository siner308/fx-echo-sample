# FX Echo Sample - 팀 데모 발표 대본

> **목표**: 팀원들에게 프로젝트의 기술적 가치와 아키텍처 우수성을 15분 내에 명확히 전달

## 🎯 발표 개요 (30초)

안녕하세요! 오늘은 제가 개발한 **FX Echo Sample** 프로젝트를 소개하겠습니다.

이 프로젝트는 **모던 Go 웹 개발의 모범사례**를 담은 RESTful API 서버로, 실제 서비스에서 바로 활용할 수 있는 **아이템 보상 시스템**을 구현했습니다.

핵심 키워드는 다음과 같습니다:
- **🏗️ Enterprise Architecture**: Uber FX + Echo + DDD
- **🔐 Security First**: Multi-JWT + Keycloak SSO + Argon2id
- **📊 API Standards**: Stripe-style Response Format

---

## 📋 발표 순서

1. **[2분]** 라이브 데모 - 실제 동작 확인
2. **[3분]** 아키텍처 개요 - 설계 철학
3. **[4분]** 핵심 기술 스택 - 기술적 차별점
4. **[3분]** 코드 워크스루 - 실제 구현
5. **[2분]** 개발 경험 - 생산성과 품질
6. **[1분]** Q&A 및 마무리

---

## 🚀 PART 1: 라이브 데모 (2분)

### 시작 멘트
> "먼저 실제로 어떻게 동작하는지 보여드리겠습니다. 새로운 팀원이 프로젝트를 받았을 때의 상황을 재현해보겠습니다."

### 데모 시퀀스

```bash
# 터미널에서 실행하며 설명
```

**1. 환경 구축 (30초)**
```bash
# 클론했다고 가정하고 시작
cd fx-echo-sample
ls -la  # 프로젝트 구조 간단히 보여주기
```

> "보시다시피 잘 정리된 프로젝트 구조를 가지고 있습니다. 이제 환경을 구축해보겠습니다."

```bash
make setup
```

> "Makefile을 통해 원클릭으로 개발 환경이 구축됩니다. .env 파일이 자동으로 생성되고 의존성이 설치됩니다."

**2. 서버 실행 (30초)**
```bash
make run
```

> "서버가 8080 포트에서 실행됩니다. Uber FX의 의존성 주입이 자동으로 이루어지는 것을 확인할 수 있습니다."

**3. API 테스트 (1분)**

새 터미널을 열고:
```bash
# 사용자 생성
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Demo User","email":"demo@example.com","age":25,"password":"password123"}'
```

> "Stripe 스타일의 깔끔한 응답을 받습니다. wrapper 없이 데이터를 직접 반환하는 것이 특징입니다."

```bash
# 로그인
curl -X POST http://localhost:8080/api/v1/auth/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"password123"}'
```

> "JWT 토큰이 발급됩니다. Access Token과 Refresh Token이 분리되어 있어 보안성이 높습니다."

```bash
# 아이템 타입 조회
curl http://localhost:8080/api/v1/items/types
```

> "비즈니스 로직이 잘 구현되어 있습니다. 6가지 아이템 타입(currency, equipment, consumable 등)을 지원합니다."

---

## 🏛️ PART 2: 아키텍처 개요 (3분)

### 화면 공유: `docs/ARCHITECTURE.md` 또는 다이어그램

**설계 철학 설명**

> "이 프로젝트의 아키텍처는 세 가지 핵심 원칙에 기반합니다."

**1. Domain-Driven Design (1분)**
```
modules/
├── auth/     # 인증 도메인
├── user/     # 사용자 관리
├── item/     # 아이템 시스템  
├── payment/  # 결제 처리
├── coupon/   # 쿠폰 시스템
└── reward/   # 보상 지급
```

> "각 모듈은 독립된 도메인으로, 명확한 책임을 가지고 있습니다. 비즈니스 로직이 기술적 관심사와 분리되어 있어 유지보수성이 뛰어납니다."

**2. Dependency Injection with Uber FX (1분)**

> "Uber FX를 사용해 의존성 주입을 구현했습니다. 이는 Google에서 사용하는 Dig 라이브러리 기반으로, 타입 안전성과 라이프사이클 관리를 보장합니다."

```go
// 예시 코드 보여주기
fx.Provide(
    jwt.NewAccessTokenService,
    jwt.NewRefreshTokenService,
    user.NewService,
)
```

> "모든 의존성이 컴파일 타임에 검증되며, 순환 종속성도 자동으로 감지됩니다."

**3. Security-First Approach (1분)**

> "보안을 최우선으로 설계했습니다:"

- **다중 JWT 토큰**: Access/Refresh/Admin 토큰 분리
- **Keycloak SSO**: 엔터프라이즈 SSO 지원
- **Argon2id 해싱**: 패스워드 보안
- **타입 안전 컨텍스트**: 컴파일 타임 안전성

---

## 🔧 PART 3: 핵심 기술 스택 (4분)

### 기술 선택의 이유

**1. Echo Framework (1분)**

> "Echo를 선택한 이유는 성능과 유연성입니다."

- **고성능**: Gin보다 유연하면서도 빠름
- **미들웨어**: API별 세밀한 인증 제어
- **타입 안전**: Context 사용 시 컴파일 타임 검증

```go
// 예시: API별 다른 미들웨어
userRouter.GET("/me", handler.GetMyInfo, userAuth.VerifyAccessToken())
adminRouter.GET("/users", handler.ListUsers, adminAuth.VerifyAdminToken())
```

**2. Uber FX vs 기존 DI (1.5분)**

> "기존의 수동 의존성 주입 대신 FX를 선택한 이유:"

**기존 방식의 문제점:**
```go
// 수동 주입 - 순서 의존적, 에러 prone
userService := user.NewService(userRepo, passwordService)
authService := auth.NewService(userService, jwtService)
handler := user.NewHandler(userService, authService) // 복잡해짐
```

**FX 방식의 장점:**
```go
// 자동 주입 - 선언적, 타입 안전
fx.Provide(
    user.NewService,     // 의존성 자동 해결
    auth.NewService,     // 순서 무관
    user.NewHandler,     // 컴파일 타임 검증
)
```

**3. API Response 표준화 (1.5분)**

> "업계 표준을 따라 API 응답을 설계했습니다. Stripe API를 벤치마킹했습니다."

**이전 (비표준):**
```json
{
  "success": true,
  "data": {"id": 1, "name": "Item"},
  "message": "Success"
}
```

**현재 (Stripe 스타일):**
```json
{"id": 1, "name": "Item"}
```

**에러 응답:**
```json
{
  "error": {
    "type": "validation_error",
    "code": "invalid_parameter", 
    "message": "Name is required",
    "param": "name"
  }
}
```

> "이렇게 하면 프론트엔드 개발자들이 훨씬 예측 가능한 API를 사용할 수 있습니다."

---

## 💻 PART 4: 코드 워크스루 (3분)

### 화면 공유: 실제 코드

**1. 모듈 구조 (1분)**

> "실제 코드를 보면서 설명하겠습니다."

`modules/user/module.go` 파일 열기:
```go
var Module = fx.Module("user",
    fx.Provide(
        repository.NewMemoryRepository,
        NewService,
        NewHandler,
    ),
    fx.Invoke(registerRoutes),
)
```

> "각 모듈은 자체적으로 완결된 구조입니다. Repository, Service, Handler가 계층적으로 구성되어 있고, 라우트 등록까지 캡슐화되어 있습니다."

**2. 타입 안전 컨텍스트 (1분)**

`modules/auth/user/context.go` 파일 열기:
```go
type ContextKey string

const (
    UserIDKey ContextKey = "user_id"
    EmailKey  ContextKey = "user_email"
)

func GetUserID(c echo.Context) (int, bool) {
    if userID, ok := c.Get(string(UserIDKey)).(int); ok {
        return userID, true
    }
    return 0, false
}
```

> "컨텍스트 키를 타입으로 정의해서 런타임 에러를 컴파일 타임에 잡을 수 있습니다. 이는 Go의 타입 시스템을 최대한 활용한 예시입니다."

**3. 에러 처리 표준화 (1분)**

`pkg/dto/error.go` 파일 열기:
```go
func NewValidationError(message, param string) ErrorResponse {
    return ErrorResponse{
        Error: ErrorDetail{
            Type:    "validation_error",
            Code:    "invalid_parameter",
            Message: message,
            Param:   param,
        },
    }
}
```

> "모든 에러가 일관된 형식을 가지도록 helper 함수들을 만들었습니다. 이렇게 하면 API 사용자가 에러를 예측 가능하게 처리할 수 있습니다."

---

## 🛠️ PART 5: 개발 경험 (3분)

### 개발 생산성

**1. 자동화된 워크플로우 (1분)**

`Makefile` 보여주기:
```bash
make setup    # 환경 구축
make run      # 개발 서버
make test     # 테스트 실행
make check    # 코드 품질 검사
make ci       # CI 파이프라인
```

> "모든 반복 작업이 자동화되어 있어 개발자가 비즈니스 로직에만 집중할 수 있습니다."

**2. 포괄적인 문서화 (1분)**

`docs/` 디렉토리 보여주기:
- `ARCHITECTURE.md` - 시스템 설계
- `FX_CONCEPTS.md` - Uber FX 학습 자료
- `SECURITY_GUIDE.md` - 보안 가이드
- `API_REFERENCE.md` - API 문서

> "신입 개발자도 빠르게 온보딩할 수 있도록 단계별 학습 자료를 준비했습니다."

**3. 테스트 전략 (1분)**

```bash
make test-cover  # 커버리지 리포트 생성
```

실행하며:
> "유닛 테스트, 통합 테스트, 벤치마크 테스트가 모두 준비되어 있습니다. 특히 Argon2id 해싱 성능을 벤치마크로 검증하고 있습니다."

### 코드 품질

```bash
make check  # 포맷팅, Vet, 린팅, 테스트 일괄 실행
```

> "코드 품질을 자동으로 검증하는 파이프라인이 구축되어 있어, 일관된 코드 스타일과 품질을 유지할 수 있습니다."

---

## 🎉 PART 6: Q&A 및 마무리 (2분)

### 핵심 가치 요약 (1분)

> "이 프로젝트의 핵심 가치를 요약하면:"

**✅ 기술적 우수성**
- 모던 Go 아키텍처 패턴 적용
- 업계 표준 API 설계
- 엔터프라이즈급 보안 구현

**✅ 실용성**
- 실제 비즈니스 로직 (아이템 보상 시스템)
- 확장 가능한 모듈 구조
- 즉시 배포 가능한 코드 품질

**✅ 개발 생산성**
- 자동화된 개발 워크플로우
- 포괄적인 문서와 학습 자료
- 신입 개발자 친화적 온보딩

### 다음 단계 제안 (30초)

> "이 프로젝트를 팀에서 활용할 수 있는 방안:"

1. **템플릿 활용**: 새 프로젝트의 보일러플레이트로 사용
2. **학습 자료**: Go 웹 개발 교육 자료로 활용
3. **코드 리뷰**: 아키텍처 패턴 학습을 위한 코드 리뷰
4. **확장 개발**: 실제 서비스 요구사항에 맞춰 확장

### 질문 받기 (30초)

> "혹시 질문이 있으시면 언제든지 말씀해 주세요."

**예상 질문과 답변:**

**Q: "왜 Gin 대신 Echo를 선택했나요?"**
A: "Gin은 성능이 좋지만 미들웨어 체인이 경직되어 있습니다. Echo는 API별로 다른 미들웨어를 적용할 수 있어 더 유연합니다."

**Q: "FX 학습 비용이 높지 않나요?"**
A: "초기 학습 비용은 있지만, 대규모 프로젝트에서는 의존성 관리가 훨씬 쉬워집니다. docs/FX_CONCEPTS.md에 단계별 학습 자료를 준비해두었습니다."

**Q: "실제 DB는 언제 연결하나요?"**
A: "Repository 인터페이스가 이미 준비되어 있어서, PostgreSQL이나 MySQL 구현체만 추가하면 됩니다. 코드 변경은 최소화됩니다."

---

## 📝 발표 팁

### 시간 관리
- **각 파트별 알람**: 폰에 2분, 5분, 9분, 12분, 14분 알람 설정
- **여유분 확보**: 실제 발표는 13분 내외로 마무리
- **Q&A 시간**: 남은 시간은 질문 답변에 활용

### 준비 사항
- **터미널 2개**: 서버 실행용 + API 테스트용
- **IDE 준비**: 주요 파일들 탭으로 미리 열어두기
- **네트워크 확인**: 데모 실패 시 스크린샷 백업
- **백업 계획**: 라이브 데모 실패 시 미리 녹화한 영상 준비

### 발표 태도
- **자신감**: 기술적 우수성에 대한 확신 표현
- **실용성 강조**: 실제 비즈니스 가치 지속적 언급
- **겸손함**: 개선점도 인정하며 피드백 요청
- **팀워크**: 팀 프로젝트로 확장할 수 있는 가능성 강조

---

## 🎬 마무리 메시지

> "이 프로젝트는 단순한 데모가 아니라, 실제 프로덕션에서 사용할 수 있는 수준의 코드 품질과 아키텍처를 가지고 있습니다. 
> 
> 팀의 기술적 역량 향상과 개발 생산성 개선에 도움이 될 것이라 확신합니다. 함께 발전시켜 나가면 좋겠습니다!"

**🎤 "감사합니다. 질문 있으시면 언제든지 말씀해 주세요!"**