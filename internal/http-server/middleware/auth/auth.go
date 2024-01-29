package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/segmentio/ksuid"
	"net/http"
	"strings"
	"time"
)

type JWTAuth struct {
	SignKey                string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	RefreshTokenCookiePath string
}

func NewJWTAuth(signKey string, accessTokenTTL, refreshTokenTTL time.Duration, refreshTokenCookiePath string) *JWTAuth {
	return &JWTAuth{
		SignKey:                signKey,
		AccessTokenTTL:         accessTokenTTL,
		RefreshTokenTTL:        refreshTokenTTL,
		RefreshTokenCookiePath: refreshTokenCookiePath,
	}
}

type TokenData struct {
	AccessToken      string
	RefreshToken     string
	Path             string
	ExpiresAt        time.Time
	HttpOnly         bool
	AdditionalFields map[string]string
}

type contextKey struct {
	name string
}

var (
	TokenCtxKey = &contextKey{"Token"}

	ErrUnauthorized             = errors.New("unauthorized")
	ErrNoTokenFound             = errors.New("no token found")
	ErrInvalidToken             = errors.New("invalid token")
	ErrUnexpectedSigningMethod  = errors.New("unexpected signing method")
	ErrNoTokenFoundInCtx        = errors.New("token not found in context")
	ErrFailedToParseTokenClaims = errors.New("failed to parse token claims from context")
)

func (j *JWTAuth) NewAccessToken(additionalClaims map[string]interface{}) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(j.AccessTokenTTL).Unix(),
	}

	if additionalClaims != nil {
		for key, value := range additionalClaims {
			claims[key] = value
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.SignKey))
}

func (j *JWTAuth) NewRefreshToken() (string, error) {
	token := ksuid.New().String()
	return token, nil
}

func Verifier(j *JWTAuth) func(http.Handler) http.Handler {
	return j.Verify(GetTokenFromHeader, GetTokenFromCookie, GetTokenFromQuery)
}

func (j *JWTAuth) Verify(findTokenFns ...func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token, err := j.FindToken(r, findTokenFns...)
			if errors.Is(err, ErrNoTokenFound) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, TokenCtxKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (j *JWTAuth) FindToken(r *http.Request, findTokenFns ...func(r *http.Request) string) (*jwt.Token, error) {
	var accessTokenString string

	for _, fn := range findTokenFns {
		accessTokenString = fn(r)
		if accessTokenString != "" {
			break
		}
	}
	if accessTokenString == "" {
		return nil, ErrNoTokenFound
	}
	return j.VerifyToken(accessTokenString)
}

func FindRefreshToken(r *http.Request) (string, error) {
	refreshToken, err := GetRefreshTokenFromHeader(r)
	if err != nil {
		return "", err
	}
	if refreshToken == "" {
		refreshToken, err = GetRefreshTokenFromCookie(r)
		if err != nil {
			return "", err
		}
	}
	if refreshToken == "" {
		return "", ErrNoTokenFound
	}
	return refreshToken, nil
}

func (j *JWTAuth) VerifyToken(accessTokenString string) (*jwt.Token, error) {
	token, err := j.DecodeToken(accessTokenString)
	if err != nil {
		return nil, Errors(err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return token, nil
}

// func (j *JWTAuth) EncodeToken

func (j *JWTAuth) DecodeToken(accessTokenString string) (*jwt.Token, error) {
	return j.ParseToken(accessTokenString)
}

func (j *JWTAuth) ParseToken(accessTokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(accessTokenString, func(token *jwt.Token) (interface{}, error) {
		// TODO: add signing method to JWTAuth
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%v: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return []byte(j.SignKey), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func Authenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := GetTokenFromContext(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if token == nil {
				http.Error(w, ErrNoTokenFound.Error(), http.StatusUnauthorized)
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func GetTokenFromHeader(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && strings.ToUpper(token[0:6]) == "BEARER" {
		return token[7:]
	} else {
		return ""
	}
}

func GetRefreshTokenFromHeader(r *http.Request) (string, error) {
	refreshToken := r.Header.Get("RefreshToken")
	if refreshToken == "" {
		// If the refreshToken is not in the headers, we try to extract it from the request body
		err := r.ParseForm()
		if err != nil {
			return "", err
		}
		refreshToken = r.FormValue("RefreshToken")
	}
	return refreshToken, nil
}

func GetTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func GetTokenFromQuery(r *http.Request) string {
	// Get token from query param named "jwt".
	return r.URL.Query().Get("jwt")
}

func GetTokenFromContext(ctx context.Context) (*jwt.Token, map[string]interface{}, error) {
	token, ok := ctx.Value(TokenCtxKey).(*jwt.Token)
	if !ok {
		return nil, nil, ErrNoTokenFoundInCtx
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, ErrFailedToParseTokenClaims
	}

	return token, claims, nil
}

func SetTokenCookie(w http.ResponseWriter, name, value, path string, expiresAt time.Time, httpOnly bool) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Expires:  expiresAt,
		HttpOnly: httpOnly,
	}
	http.SetCookie(w, &cookie)
}

func SetRefreshTokenCookie(w http.ResponseWriter, refreshToken, path string, expiresAt time.Time, httpOnly bool) {
	SetTokenCookie(w, "refreshToken", refreshToken, path, expiresAt, httpOnly)
}

func SendTokensToWeb(w http.ResponseWriter, data TokenData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	responseBody := map[string]string{"accessToken": data.AccessToken}

	if len(data.AdditionalFields) > 0 {
		for key, value := range data.AdditionalFields {
			responseBody[key] = value
		}
	}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		return
	}

	SetRefreshTokenCookie(w, data.RefreshToken, data.Path, data.ExpiresAt, data.HttpOnly)
}

func SendTokensToMobileApp(w http.ResponseWriter, data TokenData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	responseBody := map[string]string{"accessToken": data.AccessToken, "refreshToken": data.RefreshToken}

	if len(data.AdditionalFields) > 0 {
		for key, value := range data.AdditionalFields {
			responseBody[key] = value
		}
	}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		return
	}
}

func Errors(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return jwt.ErrTokenExpired
	case errors.Is(err, jwt.ErrSignatureInvalid):
		return jwt.ErrSignatureInvalid
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return jwt.ErrTokenNotValidYet
	default:
		return ErrUnauthorized
	}
}
