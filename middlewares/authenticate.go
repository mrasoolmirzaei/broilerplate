package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	conf "github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/models"
	"github.com/muety/broilerplate/services"
	"github.com/muety/broilerplate/utils"
)

const (
	// queryApiKey is the query parameter name for api key.
	queryApiKey = "api_key"
)

var (
	errEmptyKey = fmt.Errorf("the api_key is empty")
)

type AuthenticateMiddleware struct {
	config           *conf.Config
	userSrvc         services.IUserService
	optionalForPaths []string
	redirectTarget   string // optional
}

func NewAuthenticateMiddleware(userService services.IUserService) *AuthenticateMiddleware {
	return &AuthenticateMiddleware{
		config:           conf.Get(),
		userSrvc:         userService,
		optionalForPaths: []string{},
	}
}

func (m *AuthenticateMiddleware) WithOptionalFor(paths []string) *AuthenticateMiddleware {
	m.optionalForPaths = paths
	return m
}

func (m *AuthenticateMiddleware) WithRedirectTarget(path string) *AuthenticateMiddleware {
	m.redirectTarget = path
	return m
}

func (m *AuthenticateMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.ServeHTTP(w, r, h.ServeHTTP)
	})
}

func (m *AuthenticateMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var user *models.User
	user, err := m.tryGetUserByCookie(r)

	if err != nil {
		user, err = m.tryGetUserByApiKeyHeader(r)
	}
	if err != nil {
		user, err = m.tryGetUserByApiKeyQuery(r)
	}

	if err != nil || user == nil {
		if m.isOptional(r.URL.Path) {
			next(w, r)
			return
		}

		if m.redirectTarget == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(conf.ErrUnauthorized))
		} else {
			http.SetCookie(w, m.config.GetClearCookie(models.AuthCookieKey, "/"))
			http.Redirect(w, r, m.redirectTarget, http.StatusFound)
		}
		return
	}

	SetPrincipal(r, user)
	next(w, r)
}

func (m *AuthenticateMiddleware) isOptional(requestPath string) bool {
	for _, p := range m.optionalForPaths {
		if strings.HasPrefix(requestPath, p) || requestPath == p {
			return true
		}
	}
	return false
}

func (m *AuthenticateMiddleware) tryGetUserByApiKeyHeader(r *http.Request) (*models.User, error) {
	key, err := utils.ExtractBearerAuth(r)
	if err != nil {
		return nil, err
	}

	var user *models.User
	userKey := strings.TrimSpace(key)
	user, err = m.userSrvc.GetUserByKey(userKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *AuthenticateMiddleware) tryGetUserByApiKeyQuery(r *http.Request) (*models.User, error) {
	key := r.URL.Query().Get(queryApiKey)
	var user *models.User
	userKey := strings.TrimSpace(key)
	if userKey == "" {
		return nil, errEmptyKey
	}
	user, err := m.userSrvc.GetUserByKey(userKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *AuthenticateMiddleware) tryGetUserByCookie(r *http.Request) (*models.User, error) {
	username, err := utils.ExtractCookieAuth(r, m.config)
	if err != nil {
		return nil, err
	}

	user, err := m.userSrvc.GetUserById(*username)
	if err != nil {
		return nil, err
	}

	// no need to check password here, as securecookie decoding will fail anyway,
	// if cookie is not properly signed

	return user, nil
}
