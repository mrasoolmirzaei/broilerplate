package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	conf "github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/models"
	"github.com/muety/broilerplate/models/view"
	"github.com/muety/broilerplate/services"
	"net/http"
)

type HomeHandler struct {
	config       *conf.Config
	keyValueSrvc services.IKeyValueService
}

var loginDecoder = schema.NewDecoder()
var signupDecoder = schema.NewDecoder()
var resetPasswordDecoder = schema.NewDecoder()

func NewHomeHandler(keyValueService services.IKeyValueService) *HomeHandler {
	return &HomeHandler{
		config:       conf.Get(),
		keyValueSrvc: keyValueService,
	}
}

func (h *HomeHandler) RegisterRoutes(router *mux.Router) {
	router.Path("/").Methods(http.MethodGet).HandlerFunc(h.GetIndex)
}

func (h *HomeHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/dashboard", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.IndexTemplate].Execute(w, h.buildViewModel(r))
}

func (h *HomeHandler) buildViewModel(r *http.Request) *view.HomeViewModel {
	return &view.HomeViewModel{
		Success: r.URL.Query().Get("success"),
		Error:   r.URL.Query().Get("error"),
	}
}
