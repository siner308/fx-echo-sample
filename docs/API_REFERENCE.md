# API 참조 문서

이 문서는 fx-echo-sample 프로젝트의 전체 API 엔드포인트를 설명합니다.

## 기본 정보

- **베이스 URL**: `http://localhost:8080`
- **API 버전**: v1
- **인증 방식**: JWT Bearer Token

## 인증

### 토큰 타입
- **Access Token**: 일반 사용자 API 접근 (1시간 유효)
- **Refresh Token**: Access Token 갱신 (24시간 유효)  
- **Admin Token**: 관리자 API 접근 (Keycloak 또는 JWT)

### 헤더 형식
```
Authorization: Bearer <token>
```

## 인증 API

### 사용자 로그인
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**응답:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 토큰 갱신
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 관리자 로그인
```http
POST /api/v1/auth/admin/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "admin_password"
}
```

## 사용자 관리 API

### 회원가입 (인증 불필요)
```http
POST /api/v1/users/signup
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 25,
  "password": "secure_password"
}
```

### 내 정보 조회 (사용자 인증)
```http
GET /api/v1/users/me
Authorization: Bearer <access_token>
```

**응답:**
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "age": 25
}
```

### 사용자 정보 조회 (사용자 인증)
```http
GET /api/v1/users/{id}
Authorization: Bearer <access_token>
```

### 프로필 수정 (사용자 인증)
```http
PUT /api/v1/users/{id}
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "email": "updated@example.com",
  "age": 26,
  "password": "new_password" // 선택사항
}
```

### 계정 삭제 (사용자 인증)
```http
DELETE /api/v1/users/{id}
Authorization: Bearer <access_token>
```

### 전체 사용자 목록 (관리자 인증)
```http
GET /api/v1/users
Authorization: Bearer <admin_token>
```

## 아이템 관리 API

### 아이템 생성 (관리자 인증)
```http
POST /api/v1/items
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Fire Sword",
  "description": "A powerful sword with fire enchantment",
  "type": "equipment",
  "metadata": {
    "attack": 100,
    "rarity": "legendary"
  }
}
```

### 아이템 수정 (관리자 인증)
```http
PUT /api/v1/items/{id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Updated Fire Sword",
  "description": "An even more powerful sword",
  "metadata": {
    "attack": 120
  }
}
```

### 아이템 삭제 (관리자 인증)
```http
DELETE /api/v1/items/{id}
Authorization: Bearer <admin_token>
```

### 아이템 목록 (관리자 인증)
```http
GET /api/v1/items
Authorization: Bearer <admin_token>
```

### 아이템 조회 (사용자 인증)
```http
GET /api/v1/items/{id}
Authorization: Bearer <access_token>
```

### 사용자 인벤토리 조회 (사용자 인증)
```http
GET /api/v1/items/user/{userID}/inventory
Authorization: Bearer <access_token>
```

**응답:**
```json
[
  {
    "item_id": 1,
    "item_name": "Fire Sword",
    "item_type": "equipment",
    "quantity": 1,
    "metadata": {
      "attack": 100,
      "rarity": "legendary"
    }
  }
]
```

## 결제 관리 API

### 결제 생성 (사용자 인증)
```http
POST /api/v1/payments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "amount": 9.99,
  "currency": "USD",
  "payment_method": "credit_card",
  "reward_items": [
    {
      "item_id": 1,
      "quantity": 1
    }
  ]
}
```

### 결제 상태 변경 (관리자 인증)
```http
PUT /api/v1/payments/{id}/status
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "status": "completed"
}
```

### 결제 조회 (사용자 인증)
```http
GET /api/v1/payments/{id}
Authorization: Bearer <access_token>
```

### 사용자 결제 목록 (사용자 인증)
```http
GET /api/v1/payments/user/{userID}
Authorization: Bearer <access_token>
```

## 리워드 관리 API

### 리워드 생성 (관리자 인증)
```http
POST /api/v1/rewards
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Welcome Bonus",
  "description": "Starter items for new players",
  "items": [
    {
      "item_id": 1,
      "quantity": 1
    },
    {
      "item_id": 2,
      "quantity": 5
    }
  ]
}
```

### 리워드 수정 (관리자 인증)
```http
PUT /api/v1/rewards/{id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Updated Welcome Bonus",
  "items": [
    {
      "item_id": 1,
      "quantity": 2
    }
  ]
}
```

### 리워드 삭제 (관리자 인증)
```http
DELETE /api/v1/rewards/{id}
Authorization: Bearer <admin_token>
```

### 리워드 목록 (관리자 인증)
```http
GET /api/v1/rewards
Authorization: Bearer <admin_token>
```

### 리워드 지급 (관리자 인증)
```http
POST /api/v1/rewards/grant
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "user_id": 1,
  "reward_id": 1
}
```

## 쿠폰 관리 API

### 쿠폰 생성 (관리자 인증)
```http
POST /api/v1/coupons
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "code": "WELCOME2024",
  "type": "item_reward",
  "discount_percent": 0,
  "discount_amount": 0,
  "reward_items": [
    {
      "item_id": 1,
      "quantity": 1
    }
  ],
  "max_uses": 1000,
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### 쿠폰 수정 (관리자 인증)
```http
PUT /api/v1/coupons/{id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "max_uses": 2000,
  "expires_at": "2025-01-31T23:59:59Z"
}
```

### 쿠폰 삭제 (관리자 인증)
```http
DELETE /api/v1/coupons/{id}
Authorization: Bearer <admin_token>
```

### 쿠폰 목록 (관리자 인증)
```http
GET /api/v1/coupons
Authorization: Bearer <admin_token>
```

### 쿠폰 조회 (사용자 인증)
```http
GET /api/v1/coupons/{id}
Authorization: Bearer <access_token>
```

### 쿠폰 사용 (사용자 인증)
```http
POST /api/v1/coupons/redeem
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "code": "WELCOME2024",
  "user_id": 1
}
```

## 데이터 타입

### 아이템 타입
- `currency`: 게임 내 화폐 (골드, 다이아몬드 등)
- `equipment`: 장비 아이템 (무기, 방어구 등)
- `consumable`: 소모품 (포션, 버프 아이템 등)
- `card`: 수집형 카드 (트레이딩 카드, 캐릭터 카드 등)
- `material`: 제작/강화 재료 (원석, 부품 등)
- `ticket`: 이용권 (던전 입장권, 특별 이벤트 티켓 등)

### 결제 상태
- `pending`: 결제 대기 중
- `processing`: 결제 처리 중
- `completed`: 결제 완료 (자동으로 아이템 지급)
- `failed`: 결제 실패
- `cancelled`: 결제 취소
- `refunded`: 결제 환불

### 쿠폰 타입
- `discount`: 할인 쿠폰 (percent 또는 amount)
- `item_reward`: 아이템 지급 쿠폰

## 에러 응답

모든 API는 다음 형식의 에러 응답을 반환합니다:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": "Additional error details if available"
}
```

### 공통 HTTP 상태 코드
- `400`: Bad Request - 잘못된 요청 데이터
- `401`: Unauthorized - 인증 토큰 없음 또는 유효하지 않음
- `403`: Forbidden - 권한 없음
- `404`: Not Found - 리소스를 찾을 수 없음
- `409`: Conflict - 리소스 충돌 (예: 이미 존재하는 이메일)
- `500`: Internal Server Error - 서버 내부 오류