# url-query

Simple helper functionality to parse the URL query parameters into a golang struct. 

## How to use

A simple example for pagination query parameters could look like in the listing below.
With the `query` tag the parameter key name can be set and with the `default` tag default values can be set.

```go
// QueryParams sample query parameter 
type QueryParams struct {
	Start       uint32 
	PageSize    uint32 `default:"25"`
}

func HandleRequest(rw http.ResponseWriter, r *http.Request) {
	var q QueryParams
	err := query.ParseRequest(r, &q)
	// handle potential error etc.
}

type QueryParams struct {
    Nested NestedStruct
	// Numbers list of numbers with custom name in tag and 
	// default values separated by comma
    Numbers []int `query:"otherName" default:"25,64"`
}

// NestedStruct that implements query.Parser interface
type NestedStruct struct {
	Field1 string
}

func (s *NestedStruct) ParseQuery(q url.Values) error {
	// do some custom logic with the url query parameters
}
```
