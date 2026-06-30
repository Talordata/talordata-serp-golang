# TalorData Go SDK

A lightweight Go SDK for the TalorData SERP API.

Repository: [Talordata/talordata-serp-golang](https://github.com/Talordata/talordata-serp-golang)

## Install

```bash
go get github.com/Talordata/talordata-serp-golang
```

## Quick Start

```go
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
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", result)
}
```

You can also set the token once in your shell:

```bash
export TALORDATA_API_TOKEN=your_token
```

Then use the package-level helper:

```go
package main

import (
	"fmt"

	talordata "github.com/Talordata/talordata-serp-golang"
)

func main() {
	result, err := talordata.Search(map[string]interface{}{
		"engine": "google",
		"q":      "car",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", result)
}
```

## API Design

- `talordata.NewClient(...)`: create a reusable client.
- `client.Search(...)`: send a form-encoded `POST` request to `/serp/v1/request` and return the parsed result.
- `client.SearchJSON(...)`: alias of `Search(...)`.
- `client.SearchHTML(...)`: return the HTML string for `json=3`.
- `client.RawSearch(...)`: return the raw HTTP response body.

## JSON Modes

- `json=1`: returns parsed JSON data. This is the default mode used by `client.Search(...)`.
- `json=2`: returns both `html` and `json`. The SDK automatically parses `data.json` into a Go value when possible.
- `json=3`: returns HTML. The SDK unwraps the HTML string from the API response.

## Example With URL

```go
package main

import (
	"fmt"
	"os"

	talordata "github.com/Talordata/talordata-serp-golang"
)

func main() {
	client := talordata.NewClient(os.Getenv("TALORDATA_API_TOKEN"))

	result, err := client.Search(map[string]interface{}{
		"url":  "https://www.google.com/search",
		"q":    "car",
		"json": 1,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", result)
}
```

## Notes

- Auth uses the `Authorization: Bearer <token>` header.
- Requests are sent as `application/x-www-form-urlencoded`.
- Boolean params are normalized to `"1"` and `"0"` before sending.
