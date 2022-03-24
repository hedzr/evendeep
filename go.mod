module github.com/hedzr/deepcopy

go 1.11

//replace github.com/hedzr/cmdr-base => ../00.cmdr-base

//replace gopkg.in/hedzr/errors.v3 => ../05.errors

//replace github.com/hedzr/log => ../10.log

//replace github.com/hedzr/logex => ../15.logex

//replace gitlab.com/hedzr/localtest => ../29.localtest

require (
	github.com/hedzr/log v1.5.41
	gitlab.com/gopriv/localtest v1.0.1
	gopkg.in/hedzr/errors.v3 v3.0.15
)
