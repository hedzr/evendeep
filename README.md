# even-deep

A library that provides deeply per-field copying, comparing abilities.

## Features

- loosely and reasonable data-types conversions, acrossing primitives, composites and functions, with customizable
  converters/transformers
- unexported values (optional), ...
- circular references immunization
- full customizable
	- user-defined value/type converters/transformers
	- *user-defined field to field name converting rule via struct Tag* [**NOT-YET**]
- easily apply different strategies
	- basic strategies are: copy-n-merge, clone,
	- strategies per struct field:
	  `slicecopy`, `slicemerge`, `mapcopy`, `mapmerge`,
	  `omitempty` (keep if source is zero or nil), `omitnil`, `omitzero`,
	  `keepneq` (keep if not equal), `cleareq` (clear if equal), ...
- copy fields by name or ordinal
	- field to field
	- field to method, method to field
	- value to function, funtion to value
	- slice[0] to struct, struct to slice[0]
	- struct to map, map to struct

- deep series
	- deepcopy: `CopyTo()`
	- deepclone: `MakeClone()`
	- deepequal: `Equal()` [NOT YET]
	- deepdiff [NOT YET]

## Usages

`deepcopy.New`, `deepcopy.MakeClone` and `deepcopy.DeepCopy` are main entries.

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

The customized Converter can be applied on transforming the data. For more information take a look [`ValueConverter`]
and [`ValueCopier`].

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

Instead of `WithValueConverters` / `WithValueCopiers`, you might register yours once by
calling `RegisterDefaultConverters` / `RegisterDefaultCopiers`.

```go
  // a stub call for coverage
	deepcopy.RegisterDefaultCopiers()

	var dst1 string
	deepcopy.RegisterDefaultConverters(&MyTypeToStringConverter{})
	deepcopy.DeepCopy(myData, &dst1)
	if dst1 != `{
  "I": 9
}` {
		t.Fatalf("bad, got %v", dst)
	}
```

### Zero Target Fields If Equals To Source

When we compare two Struct, the target one can be clear except a field value is not equal to source field. This feature
can be used for your ORM codes: someone loads a record as a golang struct variable, and make some changes,
and `deepcopy.DeepCopy(originRec, &newRecord, deepcopy.WithORMDiffOpt)`.

The codes like:

```go
func TestExample3(t *testing.T) {
	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	var originRec = deepcopy.User{ ... }
	var newRecord deepcopy.User
	var t0 = time.Unix(0, 0)
	var expectRec = deepcopy.User{Name: "Barbara", Birthday: &t0, Attr: &deepcopy.Attr{}}

	deepcopy.DeepCopy(originRec, &newRecord)
	t.Logf("newRecord: %v", newRecord)

	newRecord.Name = "Barbara"
	deepcopy.DeepCopy(originRec, &newRecord, deepcopy.WithORMDiffOpt)
	...
	if !reflect.DeepEqual(newRecord, expectRec) {
		t.Fatalf("bad, got %v | %v", newRecord, newRecord.Birthday.Nanosecond())
	}
}
```

### Keep the target value if source empty

Sometimes we would look for a do-not-modify copier, it'll keep the target field value while the corresponding source
field is empty (zero or nil). Use `deepcopy.WithOmitEmptyOpt` in the case.

```go
func TestExample4(t *testing.T) {
	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	var originRec = deepcopy.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &deepcopy.Attr{Attrs: []string{"hello", "world"}},
		Valid:     true,
	}
	var dstRecord deepcopy.User
	var t0 = time.Unix(0, 0)
	var emptyRecord = deepcopy.User{Name: "Barbara", Birthday: &t0}
	var expectRecord = deepcopy.User{Name: "Barbara", Birthday: &t0,
		Image: []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:  &deepcopy.Attr{Attrs: []string{"hello", "world"}},
		Valid: true,
	}

	// prepare a hard copy at first
	deepcopy.DeepCopy(originRec, &dstRecord)
	t.Logf("dstRecord: %v", dstRecord)

	// now update dstRecord with the non-empty fields.
	deepcopy.DeepCopy(emptyRecord, &dstRecord, deepcopy.WithOmitEmptyOpt)
	t.Logf("dstRecord: %v", dstRecord)
	if !reflect.DeepEqual(dstRecord, expectRecord) {
		t.Fatalf("bad, got %v\nexpect: %v", dstRecord, expectRecord)
	}
}
```

### String Marshalling

While copying struct, map, slice, or other source to target string, the builtin toStringConverter will be launched. And
the default logic includes marshaling the structual source to string, typically `json.Marshal`.

The default marshaller is customizable. `RegisterStringMarshaller` and `WithStringMarshaller` do it:

```go
deepcopy.RegisterStringMarshaller(yaml.Marshal)
deepcopy.RegisterStringMarshaller(json.Marshal)
```

The preset is a wraper to `json.MarshalIndent`.

### Specify CopyMergeStrategy by struct Tag

Sample struct is:

```go
type AFT struct {
	flags     flags.Flags `copy:",cleareq"`
	converter *ValueConverter
	wouldbe   int `copy:",must,keepneq,omitzero,mapmerge"`
  ignored1 int `copy:"-"`
  ignored2 int `copy:",-"`
}
```

The available tag names are:

| Tag name           | Flags                   | Detail                |
| ------------------ | ----------------------- | --------------------- |
| `-`                | `cms.Ignore`            | field will be ignored |
| `std` (*)          | `cms.Default`           | ..                    |
| `must`             | `cms.Must`              | ..                    |
| `cleareq`          | `cms.ClearIfEqual`      |                       |
| `keepneq`          | `cms.KeepIfNotEq`       |                       |
| `clearinvalid`     | `cms.ClearIfInvalid`    |                       |
| `noomit` (*)       | `cms.NoOmit`            |                       |
| `omitempty`        | `cms.OmitIfEmpty`       |                       |
| `omitnil`          | `cms.OmitIfNil`         |                       |
| `omitzero`         | `cms.OmitIfZero`        |                       |
| `noomittarget` (*) | `cms.NoOmitTarget`      |                       |
| `omitemptytarget`  | `cms.OmitIfTargetEmpty` |                       |
| `omitniltarget`    | `cms.OmitIfTargetNil`   |                       |
| `omitzerotarget`   | `cms.OmitIfTargetZero`  |                       |
| `slicecopy`        | `cms.SliceCopy`         |                       |
| `slicecopyappend`  | `cms.SliceCopyAppend`   |                       |
| `slicemerge`       | `cms.SliceMerge`        |                       |
| `mapcopy`          | `cms.MapCopy`           |                       |
| `mapmerge`         | `cms.MapMerge`          |                       |
| ...                |                         |                       |

> `*`: the flag is on by default.

## Roadmap

These features are planning but still on ice.

- [ ] Name converting and mapping
- [ ] More builtin converters
- [ ] Handle circular pointer

Issue me if you would like it or them are put on the table.

## LICENSE

under Apache 2.0.