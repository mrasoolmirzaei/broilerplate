package config

import (
	"flag"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/configor"
	"github.com/muety/broilerplate/models"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strings"
)

const (
	defaultConfigPath = "config.yml"

	SQLDialectMysql    = "mysql"
	SQLDialectPostgres = "postgres"
	SQLDialectSqlite   = "sqlite3"

	ErrUnauthorized        = "401 unauthorized"
	ErrBadRequest          = "400 bad request"
	ErrInternalServerError = "500 internal server error"

	SimpleDateFormat     = "2006-01-02"
	SimpleDateTimeFormat = "2006-01-02 15:04:05"

	MailProviderSmtp      = "smtp"
	MailProviderMailWhale = "mailwhale"
)

var emailProviders = []string{
	MailProviderSmtp,
	MailProviderMailWhale,
}

var cfg *Config
var cFlag = flag.String("config", defaultConfigPath, "config file location")
var env string

type appConfig struct {
	AvatarURLTemplate string `yaml:"avatar_url_template" default:"https://avatars.dicebear.com/api/pixel-art-neutral/{username_hash}.svg"`
}

type securityConfig struct {
	AllowSignup   bool `yaml:"allow_signup" default:"true" env:"BROILERPLATE_ALLOW_SIGNUP"`
	ExposeMetrics bool `yaml:"expose_metrics" default:"false" env:"BROILERPLATE_EXPOSE_METRICS"`
	// this is actually a pepper (https://en.wikipedia.org/wiki/Pepper_(cryptography))
	PasswordSalt    string                     `yaml:"password_salt" default:"" env:"BROILERPLATE_PASSWORD_SALT"`
	InsecureCookies bool                       `yaml:"insecure_cookies" default:"false" env:"BROILERPLATE_INSECURE_COOKIES"`
	CookieMaxAgeSec int                        `yaml:"cookie_max_age" default:"172800" env:"BROILERPLATE_COOKIE_MAX_AGE"`
	SecureCookie    *securecookie.SecureCookie `yaml:"-"`
}

type dbConfig struct {
	Host                    string `env:"BROILERPLATE_DB_HOST"`
	Port                    uint   `env:"BROILERPLATE_DB_PORT"`
	User                    string `env:"BROILERPLATE_DB_USER"`
	Password                string `env:"BROILERPLATE_DB_PASSWORD"`
	Name                    string `default:"app_db.db" env:"BROILERPLATE_DB_NAME"`
	Dialect                 string `yaml:"-"`
	Charset                 string `default:"utf8mb4" env:"BROILERPLATE_DB_CHARSET"`
	Type                    string `yaml:"dialect" default:"sqlite3" env:"BROILERPLATE_DB_TYPE"`
	MaxConn                 uint   `yaml:"max_conn" default:"2" env:"BROILERPLATE_DB_MAX_CONNECTIONS"`
	Ssl                     bool   `default:"false" env:"BROILERPLATE_DB_SSL"`
	AutoMigrateFailSilently bool   `yaml:"automigrate_fail_silently" default:"false" env:"BROILERPLATE_DB_AUTOMIGRATE_FAIL_SILENTLY"`
}

type serverConfig struct {
	Port         int    `default:"3000" env:"BROILERPLATE_PORT"`
	ListenIpV4   string `yaml:"listen_ipv4" default:"127.0.0.1" env:"BROILERPLATE_LISTEN_IPV4"`
	ListenIpV6   string `yaml:"listen_ipv6" default:"::1" env:"BROILERPLATE_LISTEN_IPV6"`
	ListenSocket string `yaml:"listen_socket" default:"" env:"BROILERPLATE_LISTEN_SOCKET"`
	TimeoutSec   int    `yaml:"timeout_sec" default:"30" env:"BROILERPLATE_TIMEOUT_SEC"`
	BasePath     string `yaml:"base_path" default:"/" env:"BROILERPLATE_BASE_PATH"`
	PublicUrl    string `yaml:"public_url" default:"http://localhost:3000" env:"BROILERPLATE_PUBLIC_URL"`
	TlsCertPath  string `yaml:"tls_cert_path" default:"" env:"BROILERPLATE_TLS_CERT_PATH"`
	TlsKeyPath   string `yaml:"tls_key_path" default:"" env:"BROILERPLATE_TLS_KEY_PATH"`
}

type mailConfig struct {
	Enabled   bool                `env:"BROILERPLATE_MAIL_ENABLED" default:"true"`
	Provider  string              `env:"BROILERPLATE_MAIL_PROVIDER" default:"smtp"`
	MailWhale MailwhaleMailConfig `yaml:"mailwhale"`
	Smtp      SMTPMailConfig      `yaml:"smtp"`
	Sender    string              `env:"BROILERPLATE_MAIL_SENDER" yaml:"sender"`
}

type MailwhaleMailConfig struct {
	Url          string `env:"BROILERPLATE_MAIL_MAILWHALE_URL"`
	ClientId     string `yaml:"client_id" env:"BROILERPLATE_MAIL_MAILWHALE_CLIENT_ID"`
	ClientSecret string `yaml:"client_secret" env:"BROILERPLATE_MAIL_MAILWHALE_CLIENT_SECRET"`
}

type SMTPMailConfig struct {
	Host     string `env:"BROILERPLATE_MAIL_SMTP_HOST"`
	Port     uint   `env:"BROILERPLATE_MAIL_SMTP_PORT"`
	Username string `env:"BROILERPLATE_MAIL_SMTP_USER"`
	Password string `env:"BROILERPLATE_MAIL_SMTP_PASS"`
	TLS      bool   `env:"BROILERPLATE_MAIL_SMTP_TLS"`
}

type Config struct {
	Env        string `default:"dev" env:"ENVIRONMENT"`
	Version    string `yaml:"-"`
	QuickStart bool   `yaml:"quick_start" env:"BROILERPLATE_QUICK_START"`
	App        appConfig
	Security   securityConfig
	Db         dbConfig
	Server     serverConfig
	Mail       mailConfig
}

func (c *Config) CreateCookie(name, value, path string) *http.Cookie {
	return c.createCookie(name, value, path, c.Security.CookieMaxAgeSec)
}

func (c *Config) GetClearCookie(name, path string) *http.Cookie {
	return c.createCookie(name, "", path, -1)
}

func (c *Config) createCookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   !c.Security.InsecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (c *Config) IsDev() bool {
	return IsDev(c.Env)
}

func (c *Config) UseTLS() bool {
	return c.Server.TlsCertPath != "" && c.Server.TlsKeyPath != ""
}

func (c *Config) GetMigrationFunc(dbDialect string) models.MigrationFunc {
	switch dbDialect {
	default:
		return func(db *gorm.DB) error {
			if err := db.AutoMigrate(&models.User{}); err != nil && !c.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.KeyStringValue{}); err != nil && !c.Db.AutoMigrateFailSilently {
				return err
			}
			return nil
		}
	}
}

func (c *dbConfig) IsSQLite() bool {
	return c.Dialect == "sqlite3"
}

func (c *dbConfig) IsMySQL() bool {
	return c.Dialect == "mysql"
}

func (c *dbConfig) IsPostgres() bool {
	return c.Dialect == "postgres"
}

func (c *serverConfig) GetPublicUrl() string {
	return strings.TrimSuffix(c.PublicUrl, "/")
}

func (c *SMTPMailConfig) ConnStr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func IsDev(env string) bool {
	return env == "dev" || env == "development"
}

func mustReadConfigLocation() string {
	if _, err := os.Stat(*cFlag); err != nil {
		logbuch.Fatal("failed to find config file at '%s'", *cFlag)
	}
	return *cFlag
}

func resolveDbDialect(dbType string) string {
	if dbType == "cockroach" {
		return "postgres"
	}
	if dbType == "sqlite" {
		return "sqlite3"
	}
	if dbType == "mariadb" {
		return "mysql"
	}
	return dbType
}

func findString(needle string, haystack []string, defaultVal string) string {
	for _, s := range haystack {
		if s == needle {
			return s
		}
	}
	return defaultVal
}

func Set(config *Config) {
	cfg = config
}

func Get() *Config {
	return cfg
}

func Load(version string) *Config {
	config := &Config{}

	flag.Parse()

	if err := configor.New(&configor.Config{}).Load(config, mustReadConfigLocation()); err != nil {
		logbuch.Fatal("failed to read config: %v", err)
	}

	env = config.Env
	config.Version = strings.TrimSpace(version)
	config.Db.Dialect = resolveDbDialect(config.Db.Type)
	config.Security.SecureCookie = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)

	if strings.HasSuffix(config.Server.BasePath, "/") {
		config.Server.BasePath = config.Server.BasePath[:len(config.Server.BasePath)-1]
	}

	// some validation checks
	if config.Server.ListenIpV4 == "" && config.Server.ListenIpV6 == "" && config.Server.ListenSocket == "" {
		logbuch.Fatal("either of listen_ipv4 or listen_ipv6 or listen_socket must be set")
	}
	if config.Db.MaxConn <= 0 {
		logbuch.Fatal("you must allow at least one database connection")
	}
	if config.Db.MaxConn > 1 && config.Db.IsSQLite() {
		logbuch.Warn("with sqlite, only a single connection is supported") // otherwise 'PRAGMA foreign_keys=ON' would somehow have to be set for every connection in the pool
		config.Db.MaxConn = 1
	}
	if config.Mail.Provider != "" && findString(config.Mail.Provider, emailProviders, "") == "" {
		logbuch.Fatal("unknown mail provider '%s'", config.Mail.Provider)
	}

	Set(config)
	return Get()
}
