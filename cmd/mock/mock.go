package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/starkers/stack-stewart/shared"
)

func main() {
	token := "48d3cdba-f530-4c67-92d0-0fe99aa53525"
	// Read file with data
	file, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatal(err)
	}
	// create an empty var of this type
	data := shared.StackList{}
	// unmarshal the data into &data
	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatal(err)
	}

	tick := time.Tick(5 * time.Second)

	total := len(data.Stacks)
	log.Printf("found %d samples to send\n", len(data.Stacks))
	for {
		select {
		case <-tick:
			for i := 0; i < len(data.Stacks); i++ {
				log.Printf(
					"[%d/%d] %s:%s/%s-%s\n",
					i+1,
					total,
					data.Stacks[i].Agent,
					data.Stacks[i].Namespace,
					data.Stacks[i].Name,
					data.Stacks[i].Kind,
				)

				data := data.Stacks[i]

				url := "http://127.0.0.1:8080/stacks"
				err := PostStack(url, data, token)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
}

// PostStack will attempt to post a stack to the API server
func PostStack(url string, data shared.Stack, token string) error {
	client := &http.Client{}
	dataBytes := new(bytes.Buffer)
	err := json.NewEncoder(dataBytes).Encode(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s", url), dataBytes)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// add authorization header to the req
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(f))
	return nil
}
