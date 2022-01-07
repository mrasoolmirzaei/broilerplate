package main

import (
	"embed"
	_ "embed"
	"github.com/emvi/logbuch"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lpar/gzipped/v2"
	conf "github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/middlewares"
	"github.com/muety/broilerplate/migrations"
	"github.com/muety/broilerplate/repositories"
	"github.com/muety/broilerplate/routes"
	"github.com/muety/broilerplate/routes/api"
	"github.com/muety/broilerplate/services"
	"github.com/muety/broilerplate/services/mail"
	fsutils "github.com/muety/broilerplate/utils/fs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Embed version.txt
//go:embed version.txt
var version string

// Embed static files
//go:embed static
var staticFiles embed.FS

var (
	db     *gorm.DB
	config *conf.Config
)

var (
	userRepository     repositories.IUserRepository
	keyValueRepository repositories.IKeyValueRepository
)

var (
	userService     services.IUserService
	mailService     services.IMailService
	keyValueService services.IKeyValueService
)

// @title Broilerplate API
// @version 1.0
// @description REST API to interact with [Broilerplate](https://github.com/muety/broilerplate)
// @description
// @description ## Authentication
// @description Set header `Authorization` to your API Key encoded as Base64 and prefixed with `Basic`
// @description **Example:** `Basic ODY2NDhkNzQtMTljNS00NTJiLWJhMDEtZmIzZWM3MGQ0YzJmCg==`

// @contact.name Ferdinand Mütsch
// @contact.url https://github.com/muety
// @contact.email ferdinand@muetsch.io

// @license.name MIT
// @license.url https://github.com/muety/broilerplate/blob/master/LICENSE.txt

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath /api
func main() {
	config = conf.Load(version)

	// Set log level
	if config.IsDev() {
		logbuch.SetLevel(logbuch.LevelDebug)
	} else {
		logbuch.SetLevel(logbuch.LevelInfo)
	}

	// Set up GORM
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Minute,
			Colorful:      false,
			LogLevel:      logger.Silent,
		},
	)

	// Connect to database
	var err error
	db, err = gorm.Open(config.Db.GetDialector(), &gorm.Config{Logger: gormLogger})
	if config.Db.IsSQLite() {
		db.Exec("PRAGMA foreign_keys = ON;")
	}

	if config.IsDev() {
		db = db.Debug()
	}
	sqlDb, err := db.DB()
	sqlDb.SetMaxIdleConns(int(config.Db.MaxConn))
	sqlDb.SetMaxOpenConns(int(config.Db.MaxConn))
	if err != nil {
		logbuch.Error(err.Error())
		logbuch.Fatal("could not connect to database")
	}
	defer sqlDb.Close()

	// Migrate database schema
	migrations.Run(db, config)

	// Repositories
	userRepository = repositories.NewUserRepository(db)
	keyValueRepository = repositories.NewKeyValueRepository(db)

	// Services
	mailService = mail.NewMailService()
	userService = services.NewUserService(mailService, userRepository)
	keyValueService = services.NewKeyValueService(keyValueRepository)

	routes.Init()

	// API Handlers
	healthApiHandler := api.NewHealthApiHandler(db)
	metricsHandler := api.NewMetricsHandler(userService, keyValueService)

	// MVC Handlers
	homeHandler := routes.NewHomeHandler(keyValueService)
	dashboardHandler := routes.NewDashboardHandler(userService)
	loginHandler := routes.NewLoginHandler(userService, mailService)
	imprintHandler := routes.NewImprintHandler(keyValueService)

	// Setup Routers
	router := mux.NewRouter()
	rootRouter := router.PathPrefix("/").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter().StrictSlash(true)

	// https://github.com/gorilla/mux/issues/416
	router.NotFoundHandler = router.NewRoute().BuildOnly().HandlerFunc(http.NotFound).GetHandler()
	router.NotFoundHandler = middlewares.NewLoggingMiddleware(logbuch.Info, []string{
		"/assets",
		"/favicon",
		"/service-worker.js",
	})(router.NotFoundHandler)

	// Globally used middlewares
	router.Use(middlewares.NewPrincipalMiddleware())
	router.Use(middlewares.NewLoggingMiddleware(logbuch.Info, []string{"/assets", "/api/health"}))
	router.Use(handlers.RecoveryHandler())

	rootRouter.Use(middlewares.NewSecurityMiddleware())

	// Route registrations
	homeHandler.RegisterRoutes(rootRouter)
	dashboardHandler.RegisterRoutes(rootRouter)
	loginHandler.RegisterRoutes(rootRouter)
	imprintHandler.RegisterRoutes(rootRouter)

	// API route registrations
	healthApiHandler.RegisterRoutes(apiRouter)
	metricsHandler.RegisterRoutes(apiRouter)

	// Static Routes
	// https://github.com/golang/go/issues/43431
	embeddedStatic, _ := fs.Sub(staticFiles, "static")
	static := conf.ChooseFS("static", embeddedStatic)

	assetsFileServer := gzipped.FileServer(fsutils.NewExistsHttpFS(
		fsutils.NewExistsFS(static).WithCache(!config.IsDev()),
	))
	staticFileServer := http.FileServer(http.FS(
		fsutils.NeuteredFileSystem{FS: static},
	))

	router.PathPrefix("/assets").Handler(assetsFileServer)
	router.PathPrefix("/swagger-ui").Handler(staticFileServer)
	router.PathPrefix("/docs").Handler(
		middlewares.NewFileTypeFilterMiddleware([]string{".go"})(staticFileServer),
	)

	// Listen HTTP
	listen(router)
}

func listen(handler http.Handler) {
	var s4, s6, sSocket *http.Server

	// IPv4
	if config.Server.ListenIpV4 != "" {
		bindString4 := config.Server.ListenIpV4 + ":" + strconv.Itoa(config.Server.Port)
		s4 = &http.Server{
			Handler:      handler,
			Addr:         bindString4,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	// IPv6
	if config.Server.ListenIpV6 != "" {
		bindString6 := "[" + config.Server.ListenIpV6 + "]:" + strconv.Itoa(config.Server.Port)
		s6 = &http.Server{
			Handler:      handler,
			Addr:         bindString6,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	// UNIX domain socket
	if config.Server.ListenSocket != "" {
		// Remove if exists
		if _, err := os.Stat(config.Server.ListenSocket); err == nil {
			logbuch.Info("--> Removing unix socket %s", config.Server.ListenSocket)
			if err := os.Remove(config.Server.ListenSocket); err != nil {
				logbuch.Fatal(err.Error())
			}
		}
		sSocket = &http.Server{
			Handler:      handler,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	if config.UseTLS() {
		if s4 != nil {
			logbuch.Info("--> Listening for HTTPS on %s... ✅", s4.Addr)
			go func() {
				if err := s4.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
		if s6 != nil {
			logbuch.Info("--> Listening for HTTPS on %s... ✅", s6.Addr)
			go func() {
				if err := s6.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
		if sSocket != nil {
			logbuch.Info("--> Listening for HTTPS on %s... ✅", config.Server.ListenSocket)
			go func() {
				unixListener, err := net.Listen("unix", config.Server.ListenSocket)
				if err != nil {
					logbuch.Fatal(err.Error())
				}
				if err := sSocket.ServeTLS(unixListener, config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
	} else {
		if s4 != nil {
			logbuch.Info("--> Listening for HTTP on %s... ✅", s4.Addr)
			go func() {
				if err := s4.ListenAndServe(); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
		if s6 != nil {
			logbuch.Info("--> Listening for HTTP on %s... ✅", s6.Addr)
			go func() {
				if err := s6.ListenAndServe(); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
		if sSocket != nil {
			logbuch.Info("--> Listening for HTTP on %s... ✅", config.Server.ListenSocket)
			go func() {
				unixListener, err := net.Listen("unix", config.Server.ListenSocket)
				if err != nil {
					logbuch.Fatal(err.Error())
				}
				if err := sSocket.Serve(unixListener); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
	}

	<-make(chan interface{}, 1)
}
