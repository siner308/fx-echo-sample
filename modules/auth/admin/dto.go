package admin

// Legacy admin login request (deprecated - use Keycloak SSO instead)
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// Keycloak SSO flow requests/responses
type KeycloakAuthURLResponse struct {
	AuthURL string `json:"auth_url"`
}

type KeycloakCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state,omitempty"`
}

type AdminLoginResponse struct {
	AdminToken    string `json:"admin_token"`
	KeycloakToken string `json:"keycloak_token,omitempty"`
	RefreshToken  string `json:"refresh_token,omitempty"`
}