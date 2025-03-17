# url-query

Simple helper functionality to parse the [url.Values](https://pkg.go.dev/net/url#Values) into a golang struct or encode a golang struct into a url values map.

## Decoding

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
    err := query.Decode(r.Url.Query(), &q)
    // handle potential error etc.
}

type QueryParams struct {
    Nested NestedStruct
    // Numbers list of numbers with custom name in tag and 
    // default values separated by comma
    Numbers []int `query:"otherName" default:"25,64"`
}

// CustomStruct that implements query.Decoder interface
type CustomStruct struct {
    Field1 string
}

func (s *CustomStruct) DecodeQuery(q url.Values) error { 
    // do some custom logic with the url query parameters
}
```

## Encoding

The same structs can also be used to encode them into a values map.
Fields can be skipped for encoding using the `query` tag set to `"-"`.
Additionally, the `omitempty` tag can be specified separated by comma, to ignore zero values from encoding.

```go
func call(req *http.Request) (*http.Response, error) {
    q := QueryParams{
        Start: 42, // set some dummy value
    }
	
    values, err := query.Encode(q)
    if err != nil {
        // handle error
    }
	
    req.URL.RawQuery = values.Encode()
    // make HTTP call with encoded query 
}

type Omit struct {
    // Field1 ignored due to - tag
    Field1 string `query:"-"` 
    // Field2 ignored if value is zero value
    Field2 string `query;"field2,omitempty"`
}

// CustomStruct that implements query.Encoder interface
type CustomStruct struct {
    Field1 string
}

func (s *CustomStruct) EncodeValues() (url.Values, error) {
    // do some custom logic and return url.Values
}
```
