package routes

import (
	"github.com/gorilla/mux"
	conf "github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/middlewares"
	"github.com/muety/broilerplate/models/view"
	"github.com/muety/broilerplate/services"
	"net/http"
)

type DashboardHandler struct {
	config   *conf.Config
	userSrvc services.IUserService
}

func NewDashboardHandler(userService services.IUserService) *DashboardHandler {
	return &DashboardHandler{
		userSrvc: userService,
		config:   conf.Get(),
	}
}

func (h *DashboardHandler) RegisterRoutes(router *mux.Router) {
	r1 := router.PathPrefix("/dashboard").Subrouter()
	r1.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).WithRedirectTarget(defaultErrorRedirectTarget()).Handler)
	r1.Methods(http.MethodGet).HandlerFunc(h.GetIndex)
}

func (h *DashboardHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.DashboardTemplate].Execute(w, h.buildViewModel(r).WithError("unauthorized"))
		return
	}

	vm := &view.DashboardViewModel{
		User: user,
	}

	templates[conf.DashboardTemplate].Execute(w, vm)
}

func (h *DashboardHandler) buildViewModel(r *http.Request) *view.DashboardViewModel {
	return &view.DashboardViewModel{
		Success: r.URL.Query().Get("success"),
		Error:   r.URL.Query().Get("error"),
	}
}
