package mini

import (
	"fmt"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DatabaseConfig struct {
	IP           string `env:"DB_HOST" envDefault:"0.0.0.0"`
	Port         int    `env:"DB_PORT" envDefault:"5432"`
	DatabaseName string `env:"DB_NAME" envDefault:""`
	User         string `env:"DB_USER"`
	Password     string `env:"DB_PASS"`
	GormPreload  bool   `env:"DB_PRELOAD" envDefault:"false"`
}

var db *gorm.DB

func DB() *gorm.DB {
	if db == nil {
		var err error
		// read database connection info from config file
		config := DatabaseConfig{}
		if !ParseConfigTo(&config) {
			Log().Fatal().Msg("Can't load database connection config")
		}
		// create connection to database
		namePass := config.User
		if config.Password != "" {
			namePass += ":" + config.Password
		}
		connectionString := fmt.Sprintf("postgresql://%s@%s:%d/%s?sslmode=disable", namePass, config.IP, config.Port, config.DatabaseName)
		db, err = gorm.Open("postgres", connectionString)
		if err != nil {
			Log().Error().Err(err).Msg("failed to connect database")
			Log().Debug().Msg("reconnecting to database...")
			time.Sleep(5 * time.Second)
			return DB()
		}
		Log().Info().Msg("Connected to database")
		if IsDebug() {
			db.LogMode(true)
			Log().Info().Msg("Set database on debuging mode")
		}
		AddUUIDGenerateExtension(db)
		if config.GormPreload {
			db = db.Set("gorm:auto_preload", true)
			Log().Info().Msg("Set auto preload mode")
		}
	}
	if err := db.DB().Ping(); err != nil {
		Log().Error().Err(err).Msg("ping failed to database")
		Log().Debug().Msg("reconnecting to database...")
		db.Close()
		db = nil
		return DB()
	}
	return db
}

func AddUUIDGenerateExtension(db *gorm.DB) {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		Log().Fatal().Msg("Can't install extension uuid-ossp")
	}
	Log().Info().Msg("Add uuid-ossp extension")
}

func Compute(b interface{}) {
	getValue := func() (reflect.Value, reflect.Type) {
		typez := reflect.TypeOf(b)
		if typez.Kind() == reflect.Ptr || typez.Kind() == reflect.Slice {
			return reflect.ValueOf(b), typez.Elem()
		}
		return reflect.ValueOf(&b), typez
	}
	valueElem, typeElem := getValue()
	if typeElem.Kind() == reflect.Struct {
		for i := 0; i < typeElem.NumField(); i++ {
			field := typeElem.Field(i)
			computeMethod := field.Tag.Get("compute")
			if computeMethod != "" {
				method := valueElem.MethodByName(computeMethod)
				if !method.IsZero() && method.IsValid() {
					method.Call([]reflect.Value{})
				}
			}
		}
	} else if typeElem.Kind() == reflect.Slice {
		for i := 0; i < valueElem.Elem().Len(); i++ {
			item := valueElem.Elem().Index(i).Interface()
			Compute(item)
		}
	}
}
