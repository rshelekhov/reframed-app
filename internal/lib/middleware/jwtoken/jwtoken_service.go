package jwtoken

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	ssogrpc "github.com/rshelekhov/reframed/internal/clients/sso/grpc"
	"github.com/rshelekhov/reframed/internal/lib/cache"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/rshelekhov/sso-protos/gen/go/sso"
)

type TokenService struct {
	ssoClient *ssogrpc.Client
	// jwks      []*ssov1.JWK
	jwksCache *cache.Cache
	mu        sync.RWMutex
	appID     int32
}

func NewJWTokenService(ssoClient *ssogrpc.Client, appID int32) *TokenService {
	return &TokenService{
		ssoClient: ssoClient,
		jwksCache: cache.New(),
		appID:     appID,
	}
}

type TokenData struct {
	AccessToken      string
	RefreshToken     string
	Domain           string
	Path             string
	ExpiresAt        time.Time
	HTTPOnly         bool
	AdditionalFields map[string]string
}

type ContextKey struct {
	name string
}

type ClaimCTXKey string

var (
	TokenCtxKey = ContextKey{"Token"}

	ErrUnauthorized             = errors.New("unauthorized")
	ErrNoTokenFound             = errors.New("no token found")
	ErrInvalidToken             = errors.New("invalid token")
	ErrUnexpectedSigningMethod  = errors.New("unexpected signing method")
	ErrNoTokenFoundInCtx        = errors.New("token not found in context")
	ErrUserIDNotFoundInCtx      = errors.New("user id not found in context")
	ErrFailedToParseTokenClaims = errors.New("failed to parse token claims from context")
)

const (
	ContextUserID = "user_id"
	CacheJWKS     = "jwks"
)

func Verifier(j *TokenService) func(http.Handler) http.Handler {
	return j.Verify(GetTokenFromHeader, GetTokenFromCookie, GetTokenFromQuery)
}

func (j *TokenService) Verify(findTokenFns ...func(r *http.Request) string) func(http.Handler) http.Handler {
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

func (j *TokenService) FindToken(r *http.Request, findTokenFns ...func(r *http.Request) string) (*jwt.Token, error) {
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

	return j.VerifyToken(r.Context(), accessTokenString)
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

	return refreshToken, nil
}

func (j *TokenService) VerifyToken(ctx context.Context, accessTokenString string) (*jwt.Token, error) {
	token, err := j.ParseToken(ctx, accessTokenString)
	if err != nil {
		return nil, Errors(err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

func (j *TokenService) ParseToken(ctx context.Context, accessTokenString string) (*jwt.Token, error) {

	jwks, err := j.GetJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Parse the tokenData using the public key
	tokenParsed, err := jwt.Parse(accessTokenString, func(token *jwt.Token) (interface{}, error) {
		kid := token.Header["kid"].(string)

		jwk, err := getJWKByKid(jwks, kid)
		if err != nil {
			return nil, err
		}

		n, err := base64.RawURLEncoding.DecodeString(jwk.GetN())
		if err != nil {
			return nil, err
		}

		e, err := base64.RawURLEncoding.DecodeString(jwk.GetE())
		if err != nil {
			return nil, err
		}

		pubKey := &rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: int(new(big.Int).SetBytes(e).Int64()),
		}

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	return tokenParsed, nil
}

func (j *TokenService) GetJWKS(ctx context.Context) ([]*ssov1.JWK, error) {
	var jwks []*ssov1.JWK

	j.mu.RLock()
	cacheValue, found := j.jwksCache.Get(CacheJWKS)
	j.mu.RUnlock()

	if found {
		cachedJWKS, ok := cacheValue.([]*ssov1.JWK)
		if !ok {
			return nil, errors.New("invalid JWKS cache value type")
		}

		jwks = make([]*ssov1.JWK, len(cachedJWKS))
		copy(jwks, cachedJWKS)
	} else {
		jwksResponse, err := j.ssoClient.Api.GetJWKS(ctx, &ssov1.GetJWKSRequest{
			AppId: j.appID,
		})

		if err != nil {
			return nil, err
		}

		jwks = jwksResponse.GetJwks()

		j.mu.Lock()
		j.jwksCache.Set(CacheJWKS, jwks, cache.DefaultExpiration)
		j.mu.Unlock()
	}

	return jwks, nil
}

//func (j *TokenService) loadJWKSFromServer(ctx context.Context) error {
//	jwks, err := j.ssoClient.Api.GetJWKS(ctx, &ssov1.GetJWKSRequest{})
//	if err != nil {
//		return err
//	}
//
//	j.jwks = jwks.GetJwks()
//
//	return nil
//}

func getJWKByKid(jwks []*ssov1.JWK, kid string) (*ssov1.JWK, error) {
	for _, jwk := range jwks {
		if jwk.GetKid() == kid {
			return jwk, nil
		}
	}
	return nil, fmt.Errorf("JWK with kid %s not found", kid)
}

func Authenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, err := GetTokenFromContext(r.Context())
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
	}

	return ""
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
	cookie, err := r.Cookie("jwtoken")
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
	// Get token from query param named "jwtoken".
	return r.URL.Query().Get("jwtoken")
}

func GetTokenFromContext(ctx context.Context) (*jwt.Token, error) {
	token, ok := ctx.Value(TokenCtxKey).(*jwt.Token)
	if !ok {
		return nil, ErrNoTokenFoundInCtx
	}

	return token, nil
}

func GetClaimsFromToken(ctx context.Context) (map[string]interface{}, error) {
	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrFailedToParseTokenClaims
	}

	return claims, nil
}

func GetUserID(ctx context.Context) (string, error) {
	claims, err := GetClaimsFromToken(ctx)
	if err != nil {
		return "", err
	}

	userID, ok := claims[ContextUserID]
	if !ok {
		return "", ErrUserIDNotFoundInCtx
	}

	return userID.(string), nil
}

func SetTokenCookie(w http.ResponseWriter, name, value, domain, path string, expiresAt time.Time, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Domain:   domain,
		Path:     path,
		Expires:  expiresAt,
		HttpOnly: httpOnly,
	})
}

func SetRefreshTokenCookie(w http.ResponseWriter, refreshToken, domain, path string, expiresAt time.Time, httpOnly bool) {
	SetTokenCookie(w, "refreshToken", refreshToken, domain, path, expiresAt, httpOnly)
}

func SendTokensToWeb(w http.ResponseWriter, data *ssov1.TokenData, httpStatus int) {
	SetRefreshTokenCookie(w,
		data.GetRefreshToken(),
		data.GetDomain(),
		data.GetPath(),
		data.GetExpiresAt().AsTime(),
		data.GetHttpOnly(),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	responseBody := map[string]string{"accessToken": data.AccessToken}

	if len(data.AdditionalFields) > 0 {
		for key, value := range data.AdditionalFields {
			responseBody[key] = value
		}
	}

	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		return
	}
}

func SendTokensToMobileApp(w http.ResponseWriter, data *ssov1.TokenData, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

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
