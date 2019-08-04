package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	//"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/starkers/stack-stewart/shared"

	echoprometheus "github.com/0neSe7en/echo-prometheus"
	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/tidwall/buntdb"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type (
	// Config ..
	Config shared.ServerConfig
	// CustomValidator ..
	CustomValidator struct {
		validator *validator.Validate
	}

	// Headers is a generic type used for headers
	Headers map[string]interface{}
)

// Validate ..
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {

	log.Info("loading config")
	config := LoadConfig("config.yaml")

	if len(config.Agents) == 0 {
		log.Fatal("you have no agents configured.. see config.yaml")
	}

	for _, k := range config.Agents {
		log.Infof("valid agents are: %s, token: %s\n", k.Name, k.Token)
	}

	db, _ := buntdb.Open(":memory:")
	db.CreateIndex("name", "*", buntdb.IndexJSON("name"))

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Debug = false
	e.Logger.SetLevel(99) //disable json logging
	e.Use(echoprometheus.NewMetric())
	e.Use(middleware.Logger())

	// recover on errors
	e.Use(middleware.Recover())

	requireToken := InitMiddlewareTokenValidator(config)

	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/healthz", Healthz())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/stacks", GetStacks(db))
	//e.Static("/", "public")

	e.POST("/stacks", PostStack(db), requireToken)

	e.Logger.Info(e.Start(":8080"))
}

// JSONStringToStack ..
func JSONStringToStack(s string) (shared.Stack, error) {
	data := shared.Stack{}
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		log.Error(err)
		return data, err
	}
	return data, err
}

// GetStacks where we will retrieve all Stacks
func GetStacks(db *buntdb.DB) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		log.Debug("GetStacks")
		raw, err := DbGetOrdered(db, "name")
		if err != nil {
			log.Fatal(err)
		}

		if len(raw) == 0 {
			return c.String(http.StatusOK, "{}")
		}

		log.Debug("more than 0")
		data := gabs.New()
		for _, v := range raw {
			stack, err := JSONStringToStack(v)
			if err != nil {
				log.Fatal(err)
			}
			trace := GetTraceName(
				stack.Agent,
				stack.Namespace,
				stack.Name,
				stack.Kind,
			)
			_, _ = data.Set(stack, "stacks", fmt.Sprintf("%s", trace))
		}
		log.Debug(data)
		c.Response().
			Header().
			Set(
				echo.HeaderContentType,
				echo.MIMEApplicationJSONCharsetUTF8)
		return c.String(http.StatusOK, data.String())
	}
}

// DbGetOrdered ...
func DbGetOrdered(db *buntdb.DB, sortKey string) ([]string, error) {
	log.Debug("dbGetOrdered")
	var resultList []string
	log.Debug("getting ordered results for key: %s\n", sortKey)
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend(sortKey, func(key, value string) bool {
			resultList = append(resultList, value)
			return true
		})
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	return resultList, err
}

// StackToJSONString converts a stack into a json string
func StackToJSONString(s *shared.Stack) string {
	j, err := json.Marshal(s)
	JSONString := string(j)
	if err != nil {
		log.Fatal(err)
	}
	return JSONString
}

func dbUpdateWithTTL(db *buntdb.DB, key string, val string) error {
	ttl := time.Minute * 10
	err := db.Update(func(tx *buntdb.Tx) error {
		opts := &buntdb.SetOptions{Expires: true, TTL: ttl}
		_, _, err := tx.Set(key, val, opts)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	return err
}

func GetTraceName(agent, namespace, name, kind string) string {
	trace := fmt.Sprintf("%s:%s/%s-%s",
		agent,
		namespace,
		name,
		kind,
	)
	return trace
}

// PostStack where we will post a single Stack
func PostStack(db *buntdb.DB) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		input := new(shared.Stack)
		if err = c.Bind(input); err != nil {
			// if binding to the struct fails
			//log.Print(input)
			fmt.Println("BIND ERR")
			fmt.Printf("%+v\n", input)
			//log.Info("BIND ERRROR.... %v", input )
			return c.String(http.StatusBadRequest, "error binding input data")
		}
		if err = c.Validate(input); err != nil {
			// return the validation failure message as a 400
			return c.String(http.StatusBadRequest, fmt.Sprintf("%v", err))
		}
		trace := GetTraceName(
			input.Agent,
			input.Namespace,
			input.Name,
			input.Kind,
		)
		inputString := StackToJSONString(input)
		err = dbUpdateWithTTL(db, trace, inputString)
		if err != nil {
			log.Panic(err)
		}
		return c.JSON(http.StatusOK, Headers{
			"updated": trace,
		})
	}
}

// LoadConfig reads from a yaml file and tries to find to the Config struct
func LoadConfig(filename string) Config {

	configRaw, err := ioutil.ReadFile(filename)
	var configLocal Config
	if err != nil {
		fmt.Printf("%s error: %v\n", filename, err)
	}

	err = yaml.Unmarshal(configRaw, &configLocal)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("config loaded from %s\n", filename)
	return configLocal

}

// InitMiddlewareTokenValidator custom middleware to check "Bearer Token"
func InitMiddlewareTokenValidator(cfg Config) echo.MiddlewareFunc {
	// Check if we got a valid token
	// curl localhost:8080/something -v -H "Authorization: Bearer <token>"
	// Note we read valid tokens from the cfg
	midFunc := middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		// Check if we got a valid token
		//   curl localhost:8080/something -v -H "Authorization: Bearer <token>"
		Validator: func(key string, c echo.Context) (bool, error) {
			// if we header matches a configured token
			for _, agent := range cfg.Agents {
				if key == agent.Token {
					return true, nil
				}
			}
			// if no matches, middleware returns false
			return false, nil
		},
	})
	return midFunc
}

// Healthz returns the healthcheck endpoint
func Healthz() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	}
}
