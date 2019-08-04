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
			b := stack2JsonEncodedBytes(data)
			//url := "http://localhost:8080/stacks"
			url := "https://httpbin.org/post"
			token := "48d3cdba-f530-4c67-92d0-0fe99aa53525"
			req, err := http.NewRequest("POST", url, b)
			req.Header.Add("Authorization", "Bearer "+token)
			req.Header.Add("Accept", "application/json")
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				log.Println("Error on response.\n[ERROR] -", err)
			}
			fmt.Printf("%d\n%s\n%s\n", res.StatusCode, res.Header, &res.Body)
			bd, _ := ioutil.ReadAll(res.Body)
			fmt.Println(string(bd))

			fmt.Printf("\n\n\n\n\n\n")
		}
		fmt.Printf("sleeping")
		time.Sleep(time.Second * 5)
	}
}

func stack2JsonEncodedBytes(s shared.Stack) *bytes.Buffer {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(s)
	return b
}
