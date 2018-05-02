# Golang any to any value copying

This library contains functions to copy any golang value to any other type.

Any compatible conversion between slice, map and struct is supported.

### Install

```
go get github.com/RangelReale/goxcopy
```

### Examples

Struct to map:
```go
type SX struct {
    Value1 string
    Value2 string `goxcopy:"value2_changed"`
}
vsx := &SX{
    Value1: "first value",
    Value2: "second value",
}

ret, err := goxcopy.CopyToNew(vsx, reflect.TypeOf(map[string]string{}))
if err != nil {
    log.Fatal(err)
}

for mn, mv := range ret.(map[string]string) {
    fmt.Printf("%s: %s\n", mn, mv)
}
```
Output:
```
Value1: first value
value2_changed: second value
```

Map with string keys to array:
```go
vsx := map[string]float32{
    "0": 16.7,
    "1": 10.5,
    "3": 99.1,
}

ret, err := goxcopy.CopyToNew(vsx, reflect.TypeOf([]float32{}))
if err != nil {
    log.Fatal(err)
}

for mn, mv := range ret.([]float32) {
    fmt.Printf("%d: %f\n", mn, mv)
}
```
Output:
```
0: 16.700001
1: 10.500000
2: 0.000000
3: 99.099998
```

Map with string keys to array, setting an existing variable:
```go
vsx := map[string]float32{
    "0": 16.7,
    "1": 10.5,
    "3": 99.1,
}

var ret []float32

err := goxcopy.CopyToExisting(vsx, &ret)
if err != nil {
    log.Fatal(err)
}

for mn, mv := range ret {
    fmt.Printf("%d: %f\n", mn, mv)
}
```
Output:
```
0: 16.700001
1: 10.500000
2: 0.000000
3: 99.099998
```

Map to struct:
```go
type sx struct {
    Value1 string
    Value2 []string
}

vsx := map[string]interface{}{
    "Value1": "first value",
    "Value2": []string{"one", "two"},
}

ret := &sx{}

err := goxcopy.CopyToExisting(vsx, ret)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Value1: %s\n", ret.Value1)
fmt.Printf("Value2: %s\n", strings.Join(ret.Value2, ","))
```
Output:
```
Value1: first value
Value2: one,two
```


### Author

Rangel Reale (rangelspam@gmail.com) 
