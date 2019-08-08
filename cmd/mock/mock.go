package main

import (
	"encoding/json"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/starkers/stack-stewart/api"
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

	tick := time.Tick(3 * time.Second)

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
				err := api.PostStack(url, data, token)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
}
