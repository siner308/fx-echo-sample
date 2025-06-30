# Claude Code Development Notes

이 문서는 Claude Code를 사용한 개발 과정에서 중요한 변경사항과 개선점을 기록합니다.

## API 응답 형식 표준화 (2024-07-01)

### 개선 배경
기존 API 응답이 일관성이 없어 클라이언트 개발 시 혼란이 있었습니다. Stripe API 스타일을 참고하여 업계 표준에 맞는 일관된 응답 형식으로 개선했습니다.

### 변경 내용

#### 1. 에러 응답 구조 변경 (Stripe 스타일)

**이전:**
```json
{
  "success": false,
  "error": "Invalid item type",
  "details": {"field": "message"}
}
```

**현재:**
```json
{
  "error": {
    "type": "invalid_request_error",
    "code": "invalid_parameter",
    "message": "Invalid item type",
    "param": "type"
  }
}
```

#### 2. 성공 응답 단순화

**이전:**
```json
{
  "success": true,
  "data": {"id": 1, "name": "Item"},
  "message": "Created successfully"
}
```

**현재:**
```json
{
  "id": 1,
  "name": "Item"
}
```

#### 3. 리스트 응답 (Stripe 스타일)

```json
{
  "object": "list",
  "data": [{"id": 1, "name": "Item"}],
  "has_more": false,
  "total_count": 1
}
```

#### 4. 삭제 응답

```json
{
  "deleted": true,
  "id": "123"
}
```

### 새로운 Helper 함수

**에러 응답:**
- `dto.NewError(message, type)` - 일반 에러
- `dto.NewAuthError(message)` - 인증 에러
- `dto.NewNotFoundError(resource)` - 리소스 없음
- `dto.NewValidationError(message, param)` - 단일 필드 검증 에러
- `dto.NewValidationErrors(err)` - 다중 필드 검증 에러

**성공 응답:**
- 단일 객체: `c.JSON(200, data)` 직접 반환
- 리스트: `dto.NewList(items)`
- 삭제: `dto.NewEmpty(id)`

### 적용된 모듈
- ✅ `modules/item/` - 완료
- ✅ `modules/user/` - 완료  
- ✅ `modules/auth/admin/` - 완료
- ✅ `modules/auth/user/` - 완료
- 🔄 `modules/payment/` - 진행 중
- ⏳ `modules/coupon/` - 대기
- ⏳ `modules/reward/` - 대기

### 장점
1. **업계 표준 준수**: Stripe API와 유사한 구조로 개발자 친화적
2. **일관성**: 모든 엔드포인트에서 동일한 응답 패턴
3. **응답 크기 최소화**: 불필요한 wrapper 제거
4. **상세한 에러 정보**: type, code, param 등으로 에러 처리 용이

### 마이그레이션 가이드

**클라이언트 코드 수정이 필요한 부분:**

```typescript
// 이전
if (response.success) {
  const data = response.data;
} else {
  const error = response.error;
}

// 현재
if (response.error) {
  const errorType = response.error.type;
  const message = response.error.message;
} else {
  const data = response; // 직접 사용
}
```

## 테스트 개선 권장사항

### 현재 상태
- JWT 토큰 검증에 대한 통합 테스트만 존재
- HTTP 응답 형식을 검증하는 핸들러 테스트 부족

### 권장 개선사항
1. **TDD 도입**: 응답 형식 변경 시 테스트가 먼저 실패하도록
2. **핸들러 테스트**: 각 엔드포인트의 응답 형식 검증
3. **에러 케이스 테스트**: 다양한 에러 상황에 대한 응답 검증

### 예시 테스트 코드

```go
func TestCreateItem_Success(t *testing.T) {
    // Given
    req := CreateItemRequest{Name: "Test Item"}
    
    // When
    resp := httptest.NewRecorder()
    handler.CreateItem(context, req)
    
    // Then
    assert.Equal(t, http.StatusCreated, resp.Code)
    
    var item ItemResponse
    json.Unmarshal(resp.Body.Bytes(), &item)
    assert.Equal(t, "Test Item", item.Name)
    // success wrapper가 없음을 확인
    assert.NotContains(t, resp.Body.String(), "success")
}

func TestCreateItem_ValidationError(t *testing.T) {
    // Given
    req := CreateItemRequest{} // 빈 요청
    
    // When
    resp := httptest.NewRecorder()
    handler.CreateItem(context, req)
    
    // Then
    assert.Equal(t, http.StatusBadRequest, resp.Code)
    
    var errorResp dto.ErrorResponse
    json.Unmarshal(resp.Body.Bytes(), &errorResp)
    assert.Equal(t, "validation_error", errorResp.Error.Type)
    assert.Equal(t, "invalid_parameter", errorResp.Error.Code)
    assert.NotEmpty(t, errorResp.Error.Message)
}
```

## 개발 팁

### Claude Code 사용 시 주의사항
1. **응답 형식 변경**: 기존 클라이언트와의 호환성 확인 필요
2. **일괄 변경**: 모든 모듈에 일관성 있게 적용
3. **테스트 우선**: 가능하면 TDD 방식으로 개발
4. **문서 업데이트**: API 변경 시 문서도 함께 업데이트

### 효과적인 프롬프트 예시
- "Stripe API 스타일로 에러 응답 구조 변경해줘"
- "모든 성공 응답에서 wrapper 제거하고 데이터만 직접 반환해줘"
- "validation 에러를 필드별로 상세하게 반환하도록 개선해줘"

## 다음 계획

1. **남은 모듈 완성**: payment, coupon, reward 모듈 응답 형식 통일
2. **핸들러 테스트 추가**: 모든 엔드포인트에 대한 응답 형식 검증
3. **API 문서 업데이트**: 새로운 응답 형식 반영
4. **클라이언트 SDK 개발**: 표준화된 응답 형식 활용