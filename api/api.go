package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
		log.Println(err)
	}

	if err != nil {
		log.Println(err)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return err
}
