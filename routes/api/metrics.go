package api

import (
	"errors"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	conf "github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/middlewares"
	"github.com/muety/broilerplate/models"
	mm "github.com/muety/broilerplate/models/metrics"
	"github.com/muety/broilerplate/services"
	"net/http"
	"sort"
)

const (
	MetricsPrefix = "broilerplate"

	DescAdminTotalUsers = "Total number of registered users."

	DescMemAllocTotal = "Total number of bytes allocated for heap"
	DescMemSysTotal   = "Total number of bytes obtained from the OS"
	DescGoroutines    = "Total number of running goroutines"
)

type MetricsHandler struct {
	config       *conf.Config
	userSrvc     services.IUserService
	keyValueSrvc services.IKeyValueService
}

func NewMetricsHandler(userService services.IUserService, keyValueService services.IKeyValueService) *MetricsHandler {
	return &MetricsHandler{
		userSrvc:     userService,
		keyValueSrvc: keyValueService,
		config:       conf.Get(),
	}
}

func (h *MetricsHandler) RegisterRoutes(router *mux.Router) {
	if !h.config.Security.ExposeMetrics {
		return
	}

	logbuch.Info("exposing prometheus metrics under /api/metrics")

	r := router.PathPrefix("/metrics").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
}

func (h *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	reqUser := middlewares.GetPrincipal(r)
	if reqUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(conf.ErrUnauthorized))
		return
	}

	var metrics mm.Metrics

	// TODO: user metrics

	if reqUser.IsAdmin {
		if adminMetrics, err := h.getAdminMetrics(reqUser); err != nil {
			logbuch.Error("%v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		} else {
			for _, m := range *adminMetrics {
				metrics = append(metrics, m)
			}
		}
	}

	sort.Sort(metrics)

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Write([]byte(metrics.Print()))
}

func (h *MetricsHandler) getAdminMetrics(user *models.User) (*mm.Metrics, error) {
	var metrics mm.Metrics

	if !user.IsAdmin {
		return nil, errors.New("unauthorized")
	}

	totalUsers, _ := h.userSrvc.Count()

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_admin_users_total",
		Desc:   DescAdminTotalUsers,
		Value:  int(totalUsers),
		Labels: []mm.Label{},
	})

	return &metrics, nil
}
