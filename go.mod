module github.com/hedzr/deepcopy

go 1.11

//replace github.com/hedzr/cmdr-base => ../00.cmdr-base

replace gopkg.in/hedzr/errors.v3 => ../05.errors

replace github.com/hedzr/log => ../10.log

//replace github.com/hedzr/logex => ../15.logex

replace github.com/hedzr/localtest => ../29.localtest

require (
	github.com/hedzr/localtest v0.0.0-00010101000000-000000000000
	github.com/hedzr/log v1.5.28
	gopkg.in/hedzr/errors.v3 v3.0.9
)
