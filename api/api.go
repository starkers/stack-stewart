package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/starkers/stack-stewart/shared"
)

// Containers ..
type Containers struct {
	Image string `json:"image" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

// Replicas ..
type Replicas struct {
	Available int32 `json:"available" validate:"required"`
	//Desired   int32 `json:"desired" validate:"required"`
	Ready   int32 `json:"ready" validate:"required"`
	Updated int32 `json:"updated" validate:"required"`
}

// Stack ..
type Stack struct {
	Agent         string       `json:"agent" validate:"required"`
	ContainerList []Containers `json:"containers" validate:"required"`
	Kind          string       `json:"kind" validate:"required"`
	Lane          string       `json:"lane" validate:"required"`
	Name          string       `json:"name" validate:"required"`
	Namespace     string       `json:"namespace" validate:"required"`
	Replicas      Replicas     `json:"replicas"`
}

// StackList  list of stacks
type StackList struct {
	Stacks []Stack `json:"stacks"`
}

// ServerConfig to be created from config.yaml
type ServerConfig struct {
	Agents []struct {
		Name  string `yaml:"name"`
		Token string `yaml:"token"`
	} `yaml:"agents"`
}

// PostStack posts Stacks to the server
func PostStack(
	url string,
	data shared.Stack,
	token string,
) error {
	client := &http.Client{}
	dataBytes := new(bytes.Buffer)
	_ = json.NewEncoder(dataBytes).Encode(data)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s", url), dataBytes)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// add authorization header to the req
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)

	if err != nil {
		log.Error("serious issue posting.. see api.PostStack()")
		log.Error(err)
	} else {
		if resp.StatusCode == 401 {
			log.Warnf("Token %s denied Post to %s ...did you whitelist the token on the server?", token, url)
		}
		log.Debugf("posted to %s and got statusCode: %d", url, resp.StatusCode)
		// only try to close if there is something to Close()
		err = resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}
	return err
}
