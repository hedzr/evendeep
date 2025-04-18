module github.com/hedzr/evendeep

go 1.23.0

toolchain go1.23.3

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

// replace github.com/hedzr/go-errors/v2 => ../libs.errors

// replace github.com/hedzr/env => ../libs.env

// replace github.com/hedzr/is => ../libs.is

// replace github.com/hedzr/logg => ../libs.logg

// replace github.com/hedzr/logz => ../libs.logz

// replace github.com/hedzr/go-utils/v2 => ../libs.utils

// replace github.com/hedzr/go-log/v2 => ../libs.log

require (
	github.com/hedzr/is v0.7.13
	github.com/hedzr/logg v0.8.13
	gopkg.in/hedzr/errors.v3 v3.3.5
)

require (
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
)
