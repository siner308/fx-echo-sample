# Environment Setup Guide

## 📋 환경변수 설정 가이드

### 🔧 필수 환경변수

다음 환경변수들은 애플리케이션 실행에 **필수**입니다. 설정되지 않으면 애플리케이션이 시작되지 않습니다.

#### **JWT 설정**
```bash
# JWT 발행자 (모든 토큰에 공통으로 사용)
JWT_ISSUER=your-service-name

# Access Token (사용자 인증)
ACCESS_TOKEN_SECRET=your_super_secret_access_key_min_32_chars
ACCESS_TOKEN_EXPIRES=15m  # 선택사항 (기본값: 15m)

# Refresh Token (토큰 갱신)
REFRESH_TOKEN_SECRET=your_super_secret_refresh_key_min_32_chars  
REFRESH_TOKEN_EXPIRES=168h  # 선택사항 (기본값: 168h = 7일)

# Admin Token (관리자 인증)
ADMIN_TOKEN_SECRET=your_super_secret_admin_key_min_32_chars
ADMIN_TOKEN_EXPIRES=1h  # 선택사항 (기본값: 1h)
```

#### **Keycloak SSO 설정**
```bash
# Keycloak 서버 정보
KEYCLOAK_BASE_URL=https://your-keycloak-server.com
KEYCLOAK_REALM=your-realm-name
KEYCLOAK_CLIENT_ID=your-client-id
KEYCLOAK_CLIENT_SECRET=your-client-secret
KEYCLOAK_REDIRECT_URL=https://your-app.com/auth/admin/callback
```

## 🚀 설정 방법

### **1. 로컬 개발환경**
```bash
# 1. 예시 파일 복사
cp .env.example .env

# 2. 실제 값으로 수정
nano .env  # 또는 원하는 에디터 사용

# 3. 애플리케이션 실행
go run main.go
```

### **2. Docker 환경**
```dockerfile
# Dockerfile에서 환경변수 설정
ENV JWT_ISSUER=myapp
ENV ACCESS_TOKEN_SECRET=your_secret_here
# ... 기타 환경변수들
```

```bash
# 또는 docker run 시 설정
docker run -e JWT_ISSUER=myapp -e ACCESS_TOKEN_SECRET=your_secret app:latest
```

### **3. 운영환경 (Kubernetes)**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
stringData:
  JWT_ISSUER: "myapp-prod"
  ACCESS_TOKEN_SECRET: "your_production_secret"
  REFRESH_TOKEN_SECRET: "your_production_refresh_secret"
  ADMIN_TOKEN_SECRET: "your_production_admin_secret"
  KEYCLOAK_CLIENT_SECRET: "your_keycloak_secret"
```

## 🔐 보안 가이드라인

### **Secret 생성**
```bash
# 강력한 랜덤 키 생성 (32+ 문자)
openssl rand -hex 32

# 또는 
head -c 32 /dev/urandom | base64

# 예시 출력: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6
```

### **보안 체크리스트**
- [ ] 모든 SECRET은 32자리 이상의 랜덤 문자열
- [ ] 개발/운영환경에서 다른 SECRET 사용
- [ ] .env 파일은 절대 Git에 커밋하지 않음
- [ ] 운영환경에서는 환경변수나 Secret 관리 시스템 사용
- [ ] 토큰 만료시간을 적절히 설정 (보안 vs 사용성)

## 🛠️ 문제 해결

### **자주 발생하는 오류**

#### **1. Required environment variable not set**
```bash
panic: Required environment variable not set: ACCESS_TOKEN_SECRET
```
**해결방법**: 해당 환경변수를 설정하세요.
```bash
export ACCESS_TOKEN_SECRET=your_secret_here
```

#### **2. Invalid token expiration format**
```bash
ERROR: Failed to parse duration: invalid format
```
**해결방법**: 만료시간 형식을 확인하세요.
```bash
# ✅ 올바른 형식
ACCESS_TOKEN_EXPIRES=15m
REFRESH_TOKEN_EXPIRES=24h
ADMIN_TOKEN_EXPIRES=1h30m

# ❌ 잘못된 형식  
ACCESS_TOKEN_EXPIRES=15minutes
REFRESH_TOKEN_EXPIRES=1day
```

#### **3. Keycloak 연결 실패**
```bash
ERROR: Failed to connect to Keycloak
```
**확인사항**:
- KEYCLOAK_BASE_URL이 올바른지 확인
- Keycloak 서버가 실행 중인지 확인
- 네트워크 연결 상태 확인
- Client ID와 Secret이 정확한지 확인

## 📋 환경변수 참조

| 변수명 | 필수여부 | 기본값 | 설명 |
|--------|----------|--------|------|
| `JWT_ISSUER` | ✅ | - | JWT 토큰 발행자 |
| `ACCESS_TOKEN_SECRET` | ✅ | - | Access Token 암호화 키 |
| `ACCESS_TOKEN_EXPIRES` | ❌ | 15m | Access Token 만료시간 |
| `REFRESH_TOKEN_SECRET` | ✅ | - | Refresh Token 암호화 키 |
| `REFRESH_TOKEN_EXPIRES` | ❌ | 168h | Refresh Token 만료시간 |
| `ADMIN_TOKEN_SECRET` | ✅ | - | Admin Token 암호화 키 |
| `ADMIN_TOKEN_EXPIRES` | ❌ | 1h | Admin Token 만료시간 |
| `KEYCLOAK_BASE_URL` | ✅ | - | Keycloak 서버 URL |
| `KEYCLOAK_REALM` | ✅ | - | Keycloak Realm 이름 |
| `KEYCLOAK_CLIENT_ID` | ✅ | - | Keycloak Client ID |
| `KEYCLOAK_CLIENT_SECRET` | ✅ | - | Keycloak Client Secret |
| `KEYCLOAK_REDIRECT_URL` | ✅ | - | OAuth2 Redirect URL |

## 🎯 추천 설정

### **개발환경**
```bash
ACCESS_TOKEN_EXPIRES=30m    # 개발 편의를 위해 길게
REFRESH_TOKEN_EXPIRES=24h   # 하루 단위로 설정
ADMIN_TOKEN_EXPIRES=2h      # 관리자 작업 시간 고려
```

### **운영환경**
```bash
ACCESS_TOKEN_EXPIRES=15m    # 보안을 위해 짧게
REFRESH_TOKEN_EXPIRES=168h  # 일주일 단위
ADMIN_TOKEN_EXPIRES=1h      # 관리자 권한은 짧게
```