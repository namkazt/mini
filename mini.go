package mini

import (
	"math/rand"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Return new Mini instance, Use this method instead of create mini isntance directly
func New() *Mini {
	rand.Seed(time.Now().UnixNano())
	m := &Mini{}
	m.init()
	return m
}

type Mini struct {
	engine *echo.Echo
}

type AppConfig struct {
	DebugMode bool   `env:"DEBUG" envDefault:"false"`
	HttpPort  string `env:"HTTP_PORT" envDefault:"3344"`
}

var appConf = &AppConfig{}

// Parse app env for input parameters
func Env() bool {
	return ParseConfigTo(appConf)
}

func ParseConfigTo(output interface{}) bool {
	if err := env.Parse(output); err != nil {
		Log().Fatal().Err(err).Msg("Error when parse environment")
	}
	return true
}

func IsDebug() bool {
	return appConf.DebugMode
}

func (m *Mini) init() {
	m.engine = echo.New()
	m.engine.Debug = appConf.DebugMode

	logConfig := DefaultLoggerConfig
	logConfig.Output = Log()
	logConfig.Format = `${remote_ip} ${data_in_out} | ${method}:${uri} | ${status} | ${latency_human} | ${error}`
	m.engine.Use(LoggerWithConfig(logConfig))

	m.engine.Use(middleware.Recover())
}

func (m *Mini) Serve() {
	m.engine.Logger.Fatal(m.engine.Start(":" + appConf.HttpPort))
}

func (m *Mini) Echo() *echo.Echo {
	return m.engine
}

//=========================================================================
// Validator
//=========================================================================
type ServiceValidator struct {
	validator *validator.Validate
}

func (sv *ServiceValidator) Validate(i interface{}) error {
	return sv.validator.Struct(i)
}

func (m *Mini) InitValidator() {
	m.engine.Validator = &ServiceValidator{validator: validator.New()}
}
