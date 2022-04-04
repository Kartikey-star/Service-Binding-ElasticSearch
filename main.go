package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"encoding/json"
	"log"
	"os"

	"github.com/gorilla/mux"
	"github.com/nebhale/client-go/bindings"

	"github.com/elastic/go-elasticsearch/v7"
)

func main() {

	var (
		r map[string]interface{}
	)

	b := bindings.FromServiceBindingRoot()
	b = bindings.Filter(b, "elasticsearch")
	if len(b) != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Incorrect number of elastic search drivers: %d\n", len(b))
		os.Exit(1)
	}
	host, _ := bindings.Get(b[0], "host")
	port, _ := bindings.Get(b[0], "port")
	username, _ := bindings.Get(b[0], "username")
	password, _ := bindings.Get(b[0], "password")

	address := fmt.Sprintf("%v:%v", host, port)
	cfg := elasticsearch.Config{
		Addresses: []string{
			address,
		},
		Username: username,
		Password: url.QueryEscape(password),
	}
	es, err := elasticsearch.NewClient(cfg)
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
	log.Printf("Response of info:", r)
	// Build the request body.
	var buildString strings.Builder
	buildString.WriteString(`{"title" : "`)
	buildString.WriteString("Tittle One")
	buildString.WriteString(`"}`)

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      "test",
		DocumentID: strconv.Itoa(1),
		Body:       strings.NewReader(buildString.String()),
		Refresh:    "true",
	}
	// Perform the request with the client.
	res, err = req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), 1)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}

	log.Println(strings.Repeat("-", 37))

	// 3. Search for the indexed documents
	//
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": "test",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	// Perform the search request.
	res, err = es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("test"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}
	router := mux.NewRouter()

	router.HandleFunc("/status", StatusHandler).Methods("GET")
	http.ListenAndServe(":3000", router)
}
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"build": "1"}`))
}

// Client: 8.0.0-SNAPSHOT
// Server: 8.0.0-SNAPSHOT
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// [201 Created] updated; version=1
// [201 Created] updated; version=1
// -------------------------------------
// [200 OK] 2 hits; took: 5ms
//  * ID=1, map[title:Test One]
//  * ID=2, map[title:Test Two]
// =====================================
