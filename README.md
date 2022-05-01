# even-deep

![Go](https://github.com/hedzr/evendeep/workflows/Go/badge.svg)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/evendeep.svg?label=release)](https://github.com/hedzr/evendeep/releases)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/hedzr/evendeep) <!-- [![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fhedzr%2Fevendeep.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fhedzr%2Fevendeep?ref=badge_shield)
--> [![go.dev](https://img.shields.io/badge/go.dev-reference-green)](https://pkg.go.dev/github.com/hedzr/evendeep)
[![Go Report Card](https://goreportcard.com/badge/github.com/hedzr/evendeep)](https://goreportcard.com/report/github.com/hedzr/evendeep)
[![codecov](https://codecov.io/gh/hedzr/evendeep/branch/master/graph/badge.svg)](https://codecov.io/gh/hedzr/evendeep)<!--
[![Coverage Status](https://coveralls.io/repos/github/hedzr/evendeep/badge.svg?branch=master)](https://coveralls.io/github/hedzr/evendeep?branch=master)-->

Per-field copying deeply, and comparing deeply abilities.

This library is designed for making everything customizable.

## Features

- loosely and reasonable data-types conversions, acrossing primitives, composites and functions, with customizable
  converters/transformers
- unexported values (optional), ...
- circular references immunization
- full customizable
	- user-defined value/type converters/transformers
	- user-defined field to field name converting rule via struct Tag
- easily apply different strategies
	- basic strategies are: copy-n-merge, clone,
	- strategies per struct field:
	  `slicecopy`, `slicemerge`, `mapcopy`, `mapmerge`,
	  `omitempty` (keep if source is zero or nil), `omitnil`, `omitzero`,
	  `keepneq` (keep if not equal), `cleareq` (clear if equal), ...
- copy fields by name or ordinal
	- field to field
	- field to method, method to field
	- value to function (as input), funtion result to value
	- slice[0] to struct, struct to slice[0]
	- struct to map, map to struct
	- User-defined extractor/getter on various source
	- User-defined setter for struct or map target (if mapkey is string)
	- ...

- deep series
	- deepcopy: [`DeepCopy()`](https://github.com/hedzr/evendeep/blob/master/deepcopy.go#L20),
	  or [`New()`](https://github.com/hedzr/evendeep/blob/master/deepcopy.go#L110)
	- deepclone:[ `MakeClone()`](https://github.com/hedzr/evendeep/blob/master/deepcopy.go#L36)
	- deepequal: [`DeepEqual()`](https://github.com/hedzr/evendeep/blob/master/equal.go#L13)
	- deepdiff: [`DeepDiff()`](https://github.com/hedzr/evendeep/blob/master/diff.go#L13)

## History

- v0.2.50
	- first public release here.

## Usages

### deepcopy

`eventdeep.New`, `eventdeep.MakeClone` and `eventdeep.DeepCopy` are main entries.

By default, `DeepCopy()` will copy and merge source into destination object. That means, a map or a slice will be merged
deeply, same to a struct.

[`New(opts...)`](https://github.com/hedzr/evendeep/blob/master/deepcopy.go#L110) gives a most even scaleable interface
than `DeepCopy`, it returns a new `DeepCopier` different to `DefaultCopyController` and you can make call
to `DeepCopier.DeepCopy(old, new, opts...)`.

In copy-n-merge mode, copying `[2, 3]` to `[3, 7]` will get `[3, 7, 2]`.

#### Getting Started

Here is a basic sample code:

```go
func TestExample1(t *testing.T) {
timeZone, _ := time.LoadLocation("America/Phoenix")
tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
src := eventdeep.Employee2{
Base: eventdeep.Base{
Name:      "Bob",
Birthday:  &tm,
Age:       24,
EmployeID: 7,
},
Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
Image:  []byte{95, 27, 43, 66, 0, 21, 210},
Attr:   &eventdeep.Attr{Attrs: []string{"hello", "world"}},
Valid:  true,
}
var dst eventdeep.User

// direct way but no error report: eventdeep.DeepCopy(src, &dst)
c := eventdeep.New()
if err := c.CopyTo(src, &dst); err != nil {
t.Fatal(err)
}
if !reflect.DeepEqual(dst, eventdeep.User{
Name:      "Bob",
Birthday:  &tm,
Age:       24,
EmployeID: 7,
Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
Image:     []byte{95, 27, 43, 66, 0, 21, 210},
Attr:      &eventdeep.Attr{Attrs: []string{"hello", "world"}},
Valid:     true,
}) {
t.Fatalf("bad, got %v", dst)
}
}
```

#### Customizing The Field Extractor

For the unconventional deep copy, we can copy field to field via a source extractor.

You need a target struct at first.

```go
func TestStructWithSourceExtractor(t *testing.T) {
c := context.WithValue(context.TODO(), "Data", map[string]typ.Any{
"A": 12,
})

tgt := struct {
A int
}{}

evendeep.DeepCopy(c, &tgt, evendeep.WithSourceValueExtractor(func (name string) typ.Any {
if m, ok := c.Value("Data").(map[string]typ.Any); ok {
return m[name]
}
return nil
}))

if tgt.A != 12 {
t.FailNow()
}
}
```

#### Customizing The Target Setter

As a contrary, you might specify a setter to handle the setting action on copying struct and/or map.

```go
func TestStructWithTargetSetter(t *testing.T) {
type srcS struct {
A int
B bool
C string
}

src := &srcS{
A: 5,
B: true,
C: "helloString",
}
tgt := map[string]typ.Any{
"Z": "str",
}

err := evendeep.New().CopyTo(src, &tgt,
evendeep.WithTargetValueSetter(func (value *reflect.Value, sourceNames ...string) (err error) {
if value != nil {
name := "Mo" + strings.Join(sourceNames, ".")
tgt[name] = value.Interface()
}
return // ErrShouldFallback to call the evendeep standard processing
}),
)

if err != nil || tgt["MoA"] != 5 || tgt["MoB"] != true || tgt["MoC"] != "helloString" || tgt["Z"] != "str" {
t.Errorf("err: %v, tgt: %v", err, tgt)
t.FailNow()
}
}
```

NOTE that the feature supports only copying on/between struct and/or map.

If you really wanna customize the setter for primitives or others, concern to implement a ValueCopier or ValueConverter.

#### `ByOrdinal` or `ByName`

`evendeep` enumerates fields in struct/map/slice with two strategies: `ByOrdinal` and `ByName`.

1. Default `ByOrdinal` assumes the copier loops all source fields and copy them to the corresponding destination with
   the ordinal order.
2. `ByName` strategy assumes the copier loops all target fields, and try copying value from the coressponding source
   field by its name.

When a name conversion rule is defined in a struct field tag, the copier will look for the name and copy value to, even
if it's in `ByOrdinal` mode.

#### Customizing A Converter

The customized Type/Value Converter can be applied on transforming the data from source. For more information take a
look [`ValueConverter`](https://github.com/hedzr/evendeep/blob/master/cvts.go#L127)
and [`ValueCopier`](https://github.com/hedzr/evendeep/blob/master/cvts.go#L133). Its take effects on checking the value
type of target or source, or both of them.

```go
type MyType struct {
	I int
}

type MyTypeToStringConverter struct{}

// Uncomment this line if you wanna implment a ValueCopier implementation too: 
// func (c *MyTypeToStringConverter) CopyTo(ctx *eventdeep.ValueConverterContext, source, target reflect.Value) (err error) { return }

func (c *MyTypeToStringConverter) Transform(ctx *eventdeep.ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() && targetType.Kind() == reflect.String {
		var str string
		if str, err = eventdeep.FallbackToBuiltinStringMarshalling(source); err == nil {
			target = reflect.ValueOf(str)
		}
	}
	return
}

func (c *MyTypeToStringConverter) Match(params *eventdeep.Params, source, target reflect.Type) (ctx *eventdeep.ValueConverterContext, yes bool) {
	sn, sp := source.Name(), source.PkgPath()
	sk, tk := source.Kind(), target.Kind()
	if yes = sk == reflect.Struct && tk == reflect.String &&
		sn == "MyType" && sp == "github.com/hedzr/eventdeep_test"; yes {
		ctx = &eventdeep.ValueConverterContext{Params: params}
	}
	return
}

func TestExample2(t *testing.T) {
	var myData = MyType{I: 9}
	var dst string
	eventdeep.DeepCopy(myData, &dst, eventdeep.WithValueConverters(&MyTypeToStringConverter{}))
	if dst != `{
  "I": 9
}` {
		t.Fatalf("bad, got %v", dst)
	}
}
```

Instead of `WithValueConverters` / `WithValueCopiers` for each times invoking `New()`, you might register yours once by
calling `RegisterDefaultConverters` / `RegisterDefaultCopiers` into global registry.

```go
  // a stub call for coverage
eventdeep.RegisterDefaultCopiers()

var dst1 string
eventdeep.RegisterDefaultConverters(&MyTypeToStringConverter{})
eventdeep.DeepCopy(myData, &dst1)
if dst1 != `{
  "I": 9
}` {
t.Fatalf("bad, got %v", dst)
}
```

#### Zero Target Fields If Equals To Source

When we compare two Struct, the target one can be clear to zero except a field value is not equal to source field. This
feature can be used for your ORM codes: someone loads a record as a golang struct variable, and make some changes, and
invoking `eventdeep.DeepCopy(originRec, &newRecord, eventdeep.WithORMDiffOpt)`, the changes will be kept in `newRecord`
and the others unchanged fields be cleanup at last.

The codes are:

```go
func TestExample3(t *testing.T) {
timeZone, _ := time.LoadLocation("America/Phoenix")
tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
var originRec = eventdeep.User{ ... }
var newRecord eventdeep.User
var t0 = time.Unix(0, 0)
var expectRec = eventdeep.User{Name: "Barbara", Birthday: &t0, Attr: &eventdeep.Attr{}}

eventdeep.DeepCopy(originRec, &newRecord)
t.Logf("newRecord: %v", newRecord)

newRecord.Name = "Barbara"
eventdeep.DeepCopy(originRec, &newRecord, eventdeep.WithORMDiffOpt)
...
if !reflect.DeepEqual(newRecord, expectRec) {
t.Fatalf("bad, got %v | %v", newRecord, newRecord.Birthday.Nanosecond())
}
}
```

#### Keep The Target Value If Source Is Empty

Sometimes we would look for a do-not-modify copier, it'll keep the value of target fields while the corresponding source
field is empty (zero or nil). Use `eventdeep.WithOmitEmptyOpt` in the case.

```go
func TestExample4(t *testing.T) {
timeZone, _ := time.LoadLocation("America/Phoenix")
tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
var originRec = eventdeep.User{
Name:      "Bob",
Birthday:  &tm,
Age:       24,
EmployeID: 7,
Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
Image:     []byte{95, 27, 43, 66, 0, 21, 210},
Attr:      &eventdeep.Attr{Attrs: []string{"hello", "world"}},
Valid:     true,
}
var dstRecord eventdeep.User
var t0 = time.Unix(0, 0)
var emptyRecord = eventdeep.User{Name: "Barbara", Birthday: &t0}
var expectRecord = eventdeep.User{Name: "Barbara", Birthday: &t0,
Image: []byte{95, 27, 43, 66, 0, 21, 210},
Attr:  &eventdeep.Attr{Attrs: []string{"hello", "world"}},
Valid: true,
}

// prepare a hard copy at first
eventdeep.DeepCopy(originRec, &dstRecord)
t.Logf("dstRecord: %v", dstRecord)

// now update dstRecord with the non-empty fields.
eventdeep.DeepCopy(emptyRecord, &dstRecord, eventdeep.WithOmitEmptyOpt)
t.Logf("dstRecord: %v", dstRecord)
if !reflect.DeepEqual(dstRecord, expectRecord) {
t.Fatalf("bad, got %v\nexpect: %v", dstRecord, expectRecord)
}
}
```

#### String Marshalling

While copying struct, map, slice, or other source to target string, the builtin `toStringConverter` will be launched.
And the default logic includes marshaling the structual source to string, typically `json.Marshal`.

This marshaller can be customized: `RegisterStringMarshaller` and `WithStringMarshaller` enable it:

```go
eventdeep.RegisterStringMarshaller(yaml.Marshal)
eventdeep.RegisterStringMarshaller(json.Marshal)
```

The default marshaler is a wraper to `json.MarshalIndent`.

#### Specify CopyMergeStrategy via struct Tag

Sample struct is (use `copy` as key):

```go
type AFT struct {
flags     flags.Flags `copy:",cleareq"`
converter *ValueConverter
wouldbe   int `copy:",must,keepneq,omitzero,mapmerge"`
ignored1 int `copy:"-"`
ignored2 int `copy:",-"`
}
```

##### Name conversions

`copy` tag has form: `nameConversion[,strategies...]`. `nameConversion` gives a target field Name to define a name
conversion strategy, or `-` to ignore the field.

> `nameConversion` has form:
>
> - `-`: field is ignored
> - `targetName`
> - `->targetName`
> - `sourceName->targetName`
>
> Spaces besides of `->` are allowed.

Copier will check target field tag at first, and following by a source field tag checking.

You may specify converting rule at either target or source side, Copier assume the target one is prior.

**NOTE**: `nameConversion` is fully functional only for `cms.ByName` mode. It get partial work in `cms.ByOrdinal` mode (
default mode).

*TODO*: In `cms.ByOrdinal` (`*`) mode, a name converter can be applied in copying field to field.

##### Sample codes

The test gives a sample to show you how the name-conversion and member function work together:

```go
func TestStructWithNameConversions(t *testing.T) {
type srcS struct {
A int    `copy:"A1"`
B bool   `copy:"B1,std"`
C string `copy:"C1,"`
}

type dstS struct {
A1 int
B1 bool
C1 string
}

src := &srcS{A: 6, B: true, C: "hello"}
var tgt = dstS{A1: 1}

// use ByName strategy,
err := evendeep.New().CopyTo(src, &tgt, evendeep.WithByNameStrategyOpt)

if tgt.A1 != 6 || !tgt.B1 || tgt.C1 != "hello" || err != nil {
t.Fatalf("BAD COPY, tgt: %+v", tgt)
}
}
```

#### Strategy Names

The available tag names are (Almost newest, see its
in [flags/cms/copymergestrategy.go](https://github.com/hedzr/evendeep/blob/master/flags/cms/copymergestrategy.go#L23)):

| Tag name           | Flags                   | Detail                                           |
| ------------------ | ----------------------- | ------------------------------------------------ |
| `-`                | `cms.Ignore`            | field will be ignored                            |
| `std` (*)          | `cms.Default`           | reserved                                         |
| `must`             | `cms.Must`              | reserved                                         |
| `cleareq`          | `cms.ClearIfEqual`      | set zero if target equal to source               |
| `keepneq`          | `cms.KeepIfNotEq`       | don't copy source if target not equal to source  |
| `clearinvalid`     | `cms.ClearIfInvalid`    | if target field is invalid, set to zero value    |
| `noomit` (*)       | `cms.NoOmit`            |                                                  |
| `omitempty`        | `cms.OmitIfEmpty`       | if source field is empty, keep destination value |
| `omitnil`          | `cms.OmitIfNil`         |                                                  |
| `omitzero`         | `cms.OmitIfZero`        |                                                  |
| `noomittarget` (*) | `cms.NoOmitTarget`      |                                                  |
| `omitemptytarget`  | `cms.OmitIfTargetEmpty` | if target field is empty, don't copy from source |
| `omitniltarget`    | `cms.OmitIfTargetNil`   |                                                  |
| `omitzerotarget`   | `cms.OmitIfTargetZero`  |                                                  |
| `slicecopy`        | `cms.SliceCopy`         | copy elem by subscription                        |
| `slicecopyappend`  | `cms.SliceCopyAppend`   | and append more                                  |
| `slicemerge`       | `cms.SliceMerge`        | merge with order-insensitive                     |
| `mapcopy`          | `cms.MapCopy`           | copy elem by key                                 |
| `mapmerge`         | `cms.MapMerge`          | merge map deeply                                 |
| ...                |                         |                                                  |

> `*`: the flag is on by default.

#### Notes About `DeepCopy()`

Many settings are accumulated in multiple calling on `DeepCopy()`, such as `converters`, `ignoreNames`, and so on. The
underlying object is `DefaultCopyController`.

To get a fresh clean copier, `New()` or `NewFlatDeepCopier()` are the choices. BTW,
sometimes `evendeep.ResetDefaultCopyController()` might be helpful.

The only exception is copy-n-merge strategies. There flags are saved and restored on each calling on `DeepCopy()`.

#### Notes About Global Settings

Some settings are global and available to both of `DeepCopy()` and `New().CopyTo()`, such as:

1. `WithStringMarshaller` or `RegisterDefaultStringMarshaller()`
2. `RegisterDefaultConverters`
3. `RegisterDefaultCopiers`

And so on.

### deepdiff

`DeepDiff` can deeply print the differences about two objects.

```go
delta, equal := evendeep.DeepDiff([]int{3, 0, 9}, []int{9, 3, 0}, diff.WithSliceOrderedComparison(true))
t.Logf("delta: %v", delta) // ""

delta, equal := evendeep.DeepDiff([]int{3, 0}, []int{9, 3, 0}, diff.WithSliceOrderedComparison(true))
t.Logf("delta: %v", delta) // "added: [0] = 9\n"

delta, equal := evendeep.DeepDiff([]int{3, 0}, []int{9, 3, 0})
t.Logf("delta: %v", delta)
// Outputs:
//   added: [2] = <zero>
//   modified: [0] = 9 (int) (Old: 3)
//   modified: [1] = 3 (int) (Old: <zero>)

```

`DeepDiff` is a rewrote version
upon [d4l3k/messagediff]([d4l3k/messagediff at v1.2.1 (github.com)](https://github.com/d4l3k/messagediff)). This new
code enables user-defined comparer for you.

#### Ignored Names

[`diff.WithIgnoredFields(names...)`](https://github.com/hedzr/evendeep/blob/master/diff/diff.go#L41) can give a list of
names which should be ignored when comparing.

#### Slice-Order Insensitive

In normal mode, `diff` is slice-order-sensitive, that means, `[1, 2] != [2, 1]`
. [`WithSliceOrderedComparison(b bool)`](https://github.com/hedzr/evendeep/blob/master/diff/diff.go#L41) can unmind the
differences of order and as an equal.

#### Customizing Comparer

For example, `evendeep` ships a `timeComparer`:

```go
type timeComparer struct{}

func (c *timeComparer) Match(typ reflect.Type) bool {
return typ.String() == "time.Time"
}

func (c *timeComparer) Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool) {
aTime := lhs.Interface().(time.Time)
bTime := rhs.Interface().(time.Time)
if equal = aTime.Equal(bTime); !equal {
ctx.PutModified(ctx.PutPath(path), Update{Old: aTime.String(), New: bTime.String(), Typ: typfmtlite(&lhs)})
}
return
}
```

And it has been initialized into diff info struct. `timeComparer` provides a semantic comparing for `time.Time` objects.

To enable your comparer,
use [`diff.WithComparer(comparer)`](https://github.com/hedzr/evendeep/blob/master/diff/diff.go#L65).

### deepequal

Our `DeepEqual` is shortcut to `DeepDiff`:

```go
equal := evendeep.DeepEqual([]int{3, 0, 9}, []int{9, 3, 0}, diff.WithSliceOrderedComparison(true))
if !equal {
t.Errorf("expecting equal = true but got false")
}
```

For the unhandled types and objects, DeepEqual and DeepDiff will fallback to `reflect.DeepEqual()`. It's no need to
call `reflect.DeepEqual` explicitly.

## Roadmap

These features had been planning but still on ice.

- [ ] Name converting and mapping for `cms.ByOrdinal` (`*`) mode: a universal `name converter` can be applied in copying
  field to field.
- [ ] *Use SourceExtractor and TargetSetter together (might be impossible)*
- [ ] More builtin converters (*might not be a requisite*)
- [x] Handle circular pointer (DONE)

Issue me if you wanna put it or them on the table.

## LICENSE

Under Apache 2.0.