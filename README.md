# deepcopy

Yet another golang deepcopy library to provide these features:

- loosely datatypes
- widely datatypes
- full customizable
  - user-defined value converters
  - user-defined field to field name converting rule via struct Tag
- easily apply different strategies
  - basic strategies are: copy-n-merge, clone,
  - stratigies per field:
    slicecopy, slicemerge, mapcopy, mapmerge,
    omitempty (keep if source is zero or nil), omitnil, omitzero,
    omitneq (keep if not euqal), cleareq (clear if equal)
- deep series
  - deepcopy: `CopyTo()`
  - deepclone: `MakeClone()`
  - deepequal: `Equal()`
  - deepdiff [NOT YET]

## Usages

```go

```

## LICENSE

under Apache 2.0.