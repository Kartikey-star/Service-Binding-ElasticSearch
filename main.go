package main

import (
	"fmt"
	"github.com/baijum/servicebinding/binding"
	"github.com/elastic/go-elasticsearch/v7"
	"os"
	"encoding/json"
	"log"
)

func main() {
	
	var (
		r  map[string]interface{}
	)
	sb, err := binding.NewServiceBinding()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Could not read service bindings")
		os.Exit(1)
	}

	b, err := sb.Bindings("elasticsearch")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Unable to find elasticsearch binding")
		os.Exit(1)
	}
	if len(b) != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Incorrect number of Elasticsearch bindings: %d\n", len(b))
		os.Exit(1)
	}

	_, ok := b[0]["host"]
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, "No host in binding")
		os.Exit(1)
	}
	address:=fmt.Sprintf("%v:%v",b[0]["host"],b[0]["port"])
	cfg := elasticsearch.Config{
		Addresses: []string{
		  address,
		},
		Username: b[0]["user"],
		Password:b[0]["password"],
	  }
	es,err:=elasticsearch.NewClient(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	log.Printf("Response of info:",r)
	// ...
}