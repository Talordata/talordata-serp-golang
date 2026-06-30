//go:build ignore

package main

import (
	"fmt"
	"os"

	talordata "github.com/Talordata/talordata-serp-golang"
)

func main() {
	client := talordata.NewClient(os.Getenv("TALORDATA_API_TOKEN"))
	result, err := client.Search(map[string]interface{}{
		"engine": "google",
		"q":      "car",
		"json":   2,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", result)
}
