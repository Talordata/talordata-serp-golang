//go:build ignore

package main

import (
	"fmt"
	"os"

	talordata "github.com/Talordata/talordata-serp-golang"
)

func main() {
	client := talordata.NewClient(os.Getenv("TALORDATA_API_TOKEN"))
	html, err := client.SearchHTML(map[string]interface{}{
		"engine": "google",
		"q":      "car",
	})
	if err != nil {
		panic(err)
	}

	limit := 1000
	if len(html) < limit {
		limit = len(html)
	}
	fmt.Println(html[:limit])
}
