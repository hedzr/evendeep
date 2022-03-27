# deep-copy

deep-copy library provides these features:

- loosely data-types conversions, with customizable converters/transformers
- widely data-types, ...
- full customizable
  - user-defined value/type converters/transformers
  - user-defined field to field name converting rule via struct Tag
- easily apply different strategies
  - basic strategies are: copy-n-merge, clone, 
  - strategies per struct field:
    `slicecopy`, `slicemerge`, `mapcopy`, `mapmerge`,
    `omitempty` (keep if source is zero or nil), `omitnil`, `omitzero`,
    `omitneq` (keep if not equal), `cleareq` (clear if equal)
- deep series
  - deepcopy: `CopyTo()`
  - deepclone: `MakeClone()`
  - deepequal: `Equal()` [NOT YET]
  - deepdiff [NOT YET]

## Usages

```go
func TestExample1(t *testing.T) {
	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	src := deepcopy.Employee2{
		Base: deepcopy.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &deepcopy.Attr{Attrs: []string{"hello", "world"}},
		Valid:  true,
	}
	var dst deepcopy.User

  // direct way but no error report: deepcopy.DeepCopy(src, &dst)
  c := deepcopy.New()
	if err := c.CopyTo(src, &dst); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dst, deepcopy.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &deepcopy.Attr{Attrs: []string{"hello", "world"}},
		Valid:     true,
	}) {
		t.Fatalf("bad, got %v", dst)
	}
}
```



### Your Converter for A Type

You may add customized Converter and transform the data with special type transparently. For more information take a look [`ValueConverter`] and [`ValueCopier`].

```go
type MyType struct {
	I int
}

type MyTypeToStringConverter struct{}

// Uncomment this line if you wanna take a ValueCopier implementation too: 
// func (c *MyTypeToStringConverter) CopyTo(ctx *deepcopy.ValueConverterContext, source, target reflect.Value) (err error) { return }

func (c *MyTypeToStringConverter) Transform(ctx *deepcopy.ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() && targetType.Kind() == reflect.String {
		var str string
		if str, err = deepcopy.FallbackToBuiltinStringMarshalling(source); err == nil {
			target = reflect.ValueOf(str)
		}
	}
	return
}

func (c *MyTypeToStringConverter) Match(params *deepcopy.Params, source, target reflect.Type) (ctx *deepcopy.ValueConverterContext, yes bool) {
	sn, sp := source.Name(), source.PkgPath()
	sk, tk := source.Kind(), target.Kind()
	if yes = sk == reflect.Struct && tk == reflect.String &&
		sn == "MyType" && sp == "github.com/hedzr/deepcopy_test"; yes {
		ctx = &deepcopy.ValueConverterContext{Params: params}
	}
	return
}

func TestExample2(t *testing.T) {
	var myData = MyType{I: 9}
	var dst string
	deepcopy.DeepCopy(myData, &dst, deepcopy.WithValueConverters(&MyTypeToStringConverter{}))
	if dst != `{
  "I": 9
}` {
		t.Fatalf("bad, got %v", dst)
	}
}
```



### Zero Target Fields If Equals To Source

When we compare two Struct, the target one can be clear except a field value is not equal to source field. This feature can be used for your ORM codes: someone loads a record as a golang struct variable, and make some changes, and `deepcopy.DeepCopy(originRec, &newRecord, deepcopy.WithClearIfEqualOpt)`,





## LICENSE

under Apache 2.0.