package threescale

import "net/http"

type Services struct {
	Services []*Service `json:"services"`
}

type Service struct {
	Id                          int       `json:"id"`
	AccountId                   int       `json:"account_id"`
	Name                        string    `json:"name"`
	State                       string    `json:"state"`
	SystemName                  string    `json:"system_name"`
	BackendVersion              int       `json:"backend_version"`
	Description                 string    `json:"description"`
	IntentionsRequired          bool      `json:"intentions_required"`
	BuyersManageApps            bool      `json:"buyers_manage_apps"`
	BuyersManageKeys            bool      `json:"buyers_manage_keys"`
	ReferrerFiltersRequired     bool      `json:"referrer_filters_required"`
	CustomKeysEnabled           bool      `json:"custom_keys_enabled"`
	BuyerKeyRegenerateEnabled   bool      `json:"buyer_key_regenerate_enabled"`
	MandatoryAppKey             bool      `json:"mandatory_app_key"`
	BuyerCanSelectPlan          bool      `json:"buyer_can_select_plan"`
	DeploymentOption            string    `json:"deployment_option"`
	SupportEmail                string    `json:"support_email"`
	EndUserRegistrationRequired bool      `json:"end_user_registration_required"`
	Metrics                     []*Metric `json:"metrics"`
}

type Metric struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	FriendlyName string `json:"friendly_name"`
	ServiceId    int    `json:"service_id"`
	Description  string `json:"description"`
	Unit         string `json:"unit"`
}

type Users struct {
	Users []*User `json:"users"`
}

type User struct {
	UserDetails UserDetails `json:"user"`
}

type UserDetails struct {
	Id       int    `json:"id"`
	State    string `json:"state"`
	Role     string `json:"role"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AuthProviders struct {
	AuthProviders []*AuthProvider `json:"authentication_providers"`
}

type AuthProvider struct {
	ProviderDetails AuthProviderDetails `json:"authentication_provider"`
}

type AuthProviderDetails struct {
	Id                             int    `json:"id"`
	Kind                           string `json:"kind"`
	AccountType                    string `json:"account_type"`
	Name                           string `json:"name"`
	SystemName                     string `json:"system_name"`
	ClientId                       string `json:"client_id"`
	ClientSecret                   string `json:"client_secret"`
	Site                           string `json:"site"`
	AuthorizeURL                   string `json:"authorize_url"`
	SkipSSLCertificateVerification bool   `json:"skip_ssl_certificate_verification"`
	AutomaticallyApproveAccounts   bool   `json:"automatically_approve_accounts"`
	AccountId                      int    `json:"account_id"`
	UsernameKey                    string `json:"username_key"`
	IdentifierKey                  string `json:"identifier_key"`
	TrustEmail                     bool   `json:"trust_email"`
	Published                      bool   `json:"published"`
	CreatedAt                      string `json:"created_at"`
	UpdatedAt                      string `json:"updated_at"`
	CallbackUrl                    string `json:"callback_url"`
}

type tsError struct {
	message    string
	StatusCode int
}

func (tse *tsError) Error() string {
	return tse.message
}

func tsIsNotFoundError(e error) bool {
	switch e := e.(type) {
	case *tsError:
		return e.StatusCode == http.StatusNotFound
	}

	return false
}
