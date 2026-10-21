# url-query

Simple helper functionality to parse the [url.Values](https://pkg.go.dev/net/url#Values) into a golang struct or encode a golang struct into a url values map.

## Field names and types

The query parameter key of a field is its name with the first letter in lowercase, e.g. `PageSize` becomes `pageSize`.
Use the `query` tag to set a different key, e.g. `query:"page_size"`.

Supported field types are `string`, `bool`, all `int`, `uint` and `float` types, as well as pointers to and slices of these.
Types that implement [encoding.TextUnmarshaler](https://pkg.go.dev/encoding#TextUnmarshaler) or [encoding.TextMarshaler](https://pkg.go.dev/encoding#TextMarshaler) are decoded and encoded using these methods, e.g. `time.Time`, as well as pointers to and slices of them.
This takes precedence over the default handling of the underlying kind.
A `time.Duration` is decoded and encoded in the format of [time.ParseDuration](https://pkg.go.dev/time#ParseDuration), e.g. `1h30m`.
Fields of any other type are ignored, unless they implement the `query.Decoder` or `query.Encoder` interface (see below).

## Decoding

A simple example for pagination query parameters could look like in the listing below.
With the `query` tag the parameter key name can be set and with the `default` tag default values can be set.
Defaults are only used if the parameter is not part of the query.
For all types except strings, empty values like `?limit=` are treated as not provided, so the default is used.
For slices, multiple default values are separated by comma. For all other types the value is used as is, so it may contain commas.
Fields that are not slices only use the first value of the parameter if it is given multiple times.

```go
// QueryParams sample query parameter
type QueryParams struct {
    Start    uint32
    PageSize uint32 `default:"25"`
}

func HandleRequest(rw http.ResponseWriter, r *http.Request) {
    var q QueryParams
    err := query.Decode(r.URL.Query(), &q)
    // handle potential error etc.
}

// ListParams sample query parameter with a slice and a custom type
type ListParams struct {
    // Numbers list of numbers with custom name in tag and
    // default values separated by comma
    Numbers []int `query:"otherName" default:"25,64"`
    // Custom implements query.Decoder, so it is decoded using its DecodeQuery
    Custom CustomStruct
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
Additionally, the `omitempty` option can be specified separated by comma, to ignore zero values from encoding.
The key name can be left empty to use the default one, e.g. `query:",omitempty"`.

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
    Field2 string `query:"field2,omitempty"`
    // Field3 ignored if value is zero value, uses the default key "field3"
    Field3 string `query:",omitempty"`
}

// CustomStruct that implements query.Encoder interface
type CustomStruct struct {
    Field1 string
}

func (s *CustomStruct) EncodeValues() (url.Values, error) {
    // do some custom logic and return url.Values
}
```
