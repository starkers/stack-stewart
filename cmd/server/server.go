package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"errors"

	echoprometheus "github.com/0neSe7en/echo-prometheus"
	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/starkers/stack-stewart/shared"
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
	log.SetLevel(log.InfoLevel)

	log.Info("loading config.yaml")
	ServerConfig := LoadConfig("config.yaml")

	if ServerConfig.LogLevel == "debug" {
		log.Info("setting log level to Debug")
		log.SetLevel(log.DebugLevel)
	}

	if len(ServerConfig.Agents) == 0 {
		log.Fatal("you have no agents configured.. see ServerConfig.yaml")
	}

	for _, k := range ServerConfig.Agents {
		log.Infof("valid token: %s (%s)\n", k.Token, k.Name)
	}

	db, _ := buntdb.Open(":memory:")
	err := db.CreateIndex("name", "*", buntdb.IndexJSON("name"))
	if err != nil {
		log.Error(err)
	}
	//db.CreateIndex("lane", "*", buntdb.IndexJSON("lane"))

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Debug = false
	// e.Logger.SetLevel(99) //disable json logging
	e.Use(echoprometheus.NewMetric())
	// e.Use(middleware.Logger())

	// recover on errors
	e.Use(middleware.Recover())

	requireToken := InitMiddlewareTokenValidator(ServerConfig)

	e.Validator = &CustomValidator{validator: validator.New()}

	// servers the static files
	e.Static("/", "public")
	e.GET("/healthz", Healthz())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/stacks", GetStacks(db))
	e.POST("/stacks", PostStack(db, ServerConfig), requireToken)

	e.Logger.Info(e.Start(":8080"))
}

// JSONStringToStack attempts to unmarshal json into a Stack{}
func JSONStringToStack(s string) (shared.Stack, error) {
	data := shared.Stack{}
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		log.Error("unmarshal error?")
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
		// create an empty gabs JSON array
		data, err := gabs.ParseJSON([]byte(`[]`))
		c.Response().
			Header().
			Set(echo.HeaderContentType,
				echo.MIMEApplicationJSONCharsetUTF8)
		if err != nil {
			log.Fatal(err)
		}

		if len(raw) == 0 {
			log.Debug("GetStacks for 0 results from DB")
			return c.String(http.StatusOK, data.String())

		}

		log.Debugf("getstacks returned: '%s'", raw)
		for i := 0; i < len(raw); i++ {
			stringVal := raw[i]
			stack, err := JSONStringToStack(stringVal)
			if err != nil {
				log.Fatal(err)
			}
			trace := GetTraceName(
				stack.Agent,
				stack.Namespace,
				stack.Name,
				stack.Kind,
			)
			stack.Trace = trace
			traceString := fmt.Sprintf("%s", trace)
			log.Debugf("returned data for trace: %s", traceString)
			err = data.ArrayAppend(stack)
			if err != nil {
				log.Error(err)
			}

		}
		return c.String(http.StatusOK, data.String())
	}
}

// DbGetOrdered ...
func DbGetOrdered(db *buntdb.DB, sortKey string) ([]string, error) {
	log.Debug("dbGetOrdered")
	var resultList []string
	log.Debugf("getting ordered results for key: %s", sortKey)
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
func StackToJSONString(s *shared.Stack) (string, error) {
	j, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	JSONString := string(j)
	if err != nil {
		log.Fatal(err)
	}
	return JSONString, err
}

func dbUpdateWithTTL(db *buntdb.DB, key string, val string) error {
	ttl := time.Minute * 10

	if len(strings.TrimSpace(key)) == 0 {
		log.Fatal("input key is empty.. this is a bug")
		panic("woop")
	}
	if len(strings.TrimSpace(val)) == 0 {
		log.Fatal("input value is empty.. this is a bug")
		panic("woop")
	}
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

// GetTraceName will generate a unique "trace name" for each stack
func GetTraceName(agent, namespace, name, kind string) string {
	trace := fmt.Sprintf("%s:%s/%s-%s",
		agent,
		namespace,
		name,
		kind,
	)
	return trace
}

func getAgentName(token string, config Config) (string, error) {
	for _, agent := range config.Agents {
		if token == agent.Token {
			return agent.Name, nil

		}
	}
	return "no-agent-match", errors.New("no agent found")
}

// PostStack where we will post a single Stack
func PostStack(db *buntdb.DB, ServerConfig Config) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		c.Response().
			Header().
			Set(echo.HeaderContentType,
				echo.MIMEApplicationJSONCharsetUTF8)

		// parse lets find the token being used (so we can find the agents "name")
		req := c.Request()
		headers := req.Header
		inboundAuthRaw := headers.Get("Authorization")
		// need to split the field value
		inboundAuthRawSplit := strings.Split(inboundAuthRaw, " ")
		// string of the token is the second part
		tokenString := inboundAuthRawSplit[1]

		input := new(shared.Stack)
		if err = c.Bind(input); err != nil {
			// if binding to the struct fails
			log.Errorf("BIND ERR\n %+v\n", input)
			return c.String(http.StatusBadRequest, "{'error':'binding input data'}")
		}
		if err = c.Validate(input); err != nil {
			// return the validation failure message as a 400
			log.Errorf("Error validating input from %s.. : %v", tokenString, err)
			return c.String(http.StatusBadRequest, fmt.Sprintf("{'error':'%v'}", err))
		}

		agentName, err := getAgentName(tokenString, ServerConfig)
		if err != nil {
			log.Error(err)
		}
		log.Debugf("======== %s ======", agentName)

		input.Agent = agentName
		trace := GetTraceName(
			input.Agent,
			input.Namespace,
			input.Name,
			input.Kind,
		)
		inputString, err := StackToJSONString(input)
		if err != nil {
			log.Fatal(err)
		}
		err = dbUpdateWithTTL(db, trace, inputString)
		if err != nil {
			log.Fatal(err)
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
		log.Errorf("%s error: %v\n", filename, err)
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
		return c.String(http.StatusOK, "{'status':'ok'}")
	}
}
