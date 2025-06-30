package security

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		config   *PasswordConfig
		wantErr  bool
	}{
		{
			name:     "valid password with default config",
			password: "password123",
			config:   nil,
			wantErr:  false,
		},
		{
			name:     "valid password with custom config",
			password: "test123",
			config: &PasswordConfig{
				Memory:      32 * 1024,
				Iterations:  2,
				Parallelism: 1,
				SaltLength:  8,
				KeyLength:   16,
			},
			wantErr: false,
		},
		{
			name:     "empty password",
			password: "",
			config:   nil,
			wantErr:  false, // 빈 패스워드도 해싱 가능
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 1000),
			config:   nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 해시가 올바른 형식인지 확인
				if !strings.HasPrefix(hash, "$argon2id$") {
					t.Errorf("HashPassword() = %v, expected hash to start with $argon2id$", hash)
				}

				// 해시가 각기 다른지 확인 (동일한 패스워드라도 솔트가 다르므로)
				hash2, err := HashPassword(tt.password, tt.config)
				if err != nil {
					t.Errorf("HashPassword() second call error = %v", err)
				}
				if hash == hash2 {
					t.Errorf("HashPassword() generated same hash twice, expected different hashes due to random salt")
				}
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testPassword123"
	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "wrong password",
			password: "wrongPassword",
			hash:     hash,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "empty password against non-empty hash",
			password: "",
			hash:     hash,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "invalid hash format",
			password: password,
			hash:     "invalid_hash",
			want:     false,
			wantErr:  true,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
			wantErr:  true,
		},
		{
			name:     "malformed hash parts",
			password: password,
			hash:     "$argon2id$v=19$m=65536",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyPassword(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordConfigDefaults(t *testing.T) {
	config := DefaultPasswordConfig()

	if config.Memory != 64*1024 {
		t.Errorf("DefaultPasswordConfig().Memory = %v, want %v", config.Memory, 64*1024)
	}
	if config.Iterations != 3 {
		t.Errorf("DefaultPasswordConfig().Iterations = %v, want %v", config.Iterations, 3)
	}
	if config.Parallelism != 2 {
		t.Errorf("DefaultPasswordConfig().Parallelism = %v, want %v", config.Parallelism, 2)
	}
	if config.SaltLength != 16 {
		t.Errorf("DefaultPasswordConfig().SaltLength = %v, want %v", config.SaltLength, 16)
	}
	if config.KeyLength != 32 {
		t.Errorf("DefaultPasswordConfig().KeyLength = %v, want %v", config.KeyLength, 32)
	}
}

func TestDecodeHash(t *testing.T) {
	// 유효한 해시 생성
	password := "test123"
	validHash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("Failed to generate valid hash: %v", err)
	}

	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid hash",
			hash:    validHash,
			wantErr: false,
		},
		{
			name:    "invalid format - too few parts",
			hash:    "$argon2id$v=19",
			wantErr: true,
		},
		{
			name:    "invalid version format",
			hash:    "$argon2id$v=abc$m=65536,t=3,p=2$salt$hash",
			wantErr: true,
		},
		{
			name:    "invalid parameters format",
			hash:    "$argon2id$v=19$m=abc,t=3,p=2$salt$hash",
			wantErr: true,
		},
		{
			name:    "invalid base64 salt",
			hash:    "$argon2id$v=19$m=65536,t=3,p=2$invalid!!!$hash",
			wantErr: true,
		},
		{
			name:    "invalid base64 hash",
			hash:    "$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$invalid!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := decodeHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name   string
		length uint32
	}{
		{"length 0", 0},
		{"length 1", 1},
		{"length 16", 16},
		{"length 32", 32},
		{"length 64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes1, err := generateRandomBytes(tt.length)
			if err != nil {
				t.Errorf("generateRandomBytes() error = %v", err)
				return
			}

			if uint32(len(bytes1)) != tt.length {
				t.Errorf("generateRandomBytes() length = %v, want %v", len(bytes1), tt.length)
			}

			if tt.length > 0 {
				// 두 번 호출해서 다른 값인지 확인 (randomness test)
				bytes2, err := generateRandomBytes(tt.length)
				if err != nil {
					t.Errorf("generateRandomBytes() second call error = %v", err)
					return
				}

				// 아주 작은 확률로 동일할 수 있지만 일반적으로는 달라야 함
				if tt.length > 4 && string(bytes1) == string(bytes2) {
					t.Errorf("generateRandomBytes() generated same bytes twice, expected randomness")
				}
			}
		})
	}
}

// 벤치마크 테스트
func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"
	config := DefaultPasswordConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password, config)
		if err != nil {
			b.Fatalf("HashPassword() error = %v", err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkPassword123"
	hash, err := HashPassword(password, nil)
	if err != nil {
		b.Fatalf("Failed to hash password for benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := VerifyPassword(password, hash)
		if err != nil {
			b.Fatalf("VerifyPassword() error = %v", err)
		}
	}
}

// 타이밍 공격 방지 테스트
func TestConstantTimeComparison(t *testing.T) {
	password := "testPassword"
	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// 올바른 패스워드와 틀린 패스워드의 검증 시간이 일정해야 함
	// (실제로는 매우 정밀한 측정이 필요하지만, 기본적인 동작 확인)
	
	correctResults := make([]bool, 10)
	wrongResults := make([]bool, 10)

	for i := 0; i < 10; i++ {
		correctResults[i], err = VerifyPassword(password, hash)
		if err != nil {
			t.Fatalf("VerifyPassword() error = %v", err)
		}

		wrongResults[i], err = VerifyPassword("wrongPassword", hash)
		if err != nil {
			t.Fatalf("VerifyPassword() error = %v", err)
		}
	}

	// 모든 correct는 true여야 함
	for i, result := range correctResults {
		if !result {
			t.Errorf("correctResults[%d] = %v, want true", i, result)
		}
	}

	// 모든 wrong은 false여야 함
	for i, result := range wrongResults {
		if result {
			t.Errorf("wrongResults[%d] = %v, want false", i, result)
		}
	}
}