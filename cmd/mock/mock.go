package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(data)
			url := "http://localhost:8080/stacks"
			// url := "https://httpbin.org/post"
			res, err := http.Post(url, "application/json; charset=utf-8", b)
			if err != nil {
				log.Printf("%v\n", err)
			}
			io.Copy(os.Stdout, res.Body)
			fmt.Printf("\n")
		}
		fmt.Printf("sleeping")
		time.Sleep(time.Second * 5)
	}
}
