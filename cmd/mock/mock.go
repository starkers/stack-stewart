package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/starkers/stack-stewart/shared"
)

func main() {
	token := "48d3cdba-f530-4c67-92d0-0fe99aa53525"
	// Read file with data
	file, _ := ioutil.ReadFile("data.json")
	// create an empty var of this type
	data := shared.StackList{}
	// unmarshal the data into &data
	_ = json.Unmarshal([]byte(file), &data)

	total := len(data.Stacks)
	log.Printf("found %d samples to send\n", len(data.Stacks))
	for {
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

			url := "http://localhost:8080/stacks"
			_ = PostStack(url, data, token)

		}
		fmt.Printf("sleeping")
		time.Sleep(time.Second * 5)
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

	req, err := http.NewRequest("POST",fmt.Sprintf("%s", url), dataBytes)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// add authorization header to the req
	req.Header.Add("Authorization", "Bearer " + token)
	if err != nil {
		log.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(f))
	return nil
}
