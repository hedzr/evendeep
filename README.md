# deep-copy

Yet another golang deep-copy library to provide these features:

- loosely data-types conversions, with customizable converters/transformers
- widely data-types, includes chan, func, ...
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

```

## LICENSE

under Apache 2.0.