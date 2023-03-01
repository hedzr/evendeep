//go:build ignore
// +build ignore

package tests_test

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/hedzr/evendeep"
	"github.com/hedzr/evendeep/flags/cms"
)

func TestComplexStructure(t *testing.T) {
	type A struct {
		A bool
	}
	var from = A{}
	var to = A{true}
	var c = evendeep.New(
		evendeep.WithCopyStrategyOpt,
		evendeep.WithStrategies(cms.OmitIfEmpty),
	)
	var err = c.CopyTo(from, &to)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if to.A == false {
		t.Fatalf("expecting %v", true)
	}
}

func TestComplexStructure9(t *testing.T) {
	const src = `---
app:
  bgo:
    build:
      for:
        - "linux/amd64"
        - "windows/amd64"
        - "darwin/amd64"
        - "darwin/arm64"
      os: [ linux ]
      arch: [ amd64, arm64 ]
      reduce: true
      upx:
        enable: false
        params: [ "-9" ]

      projects:
        000-default-group:
          leading-text:
          items:
            003-generic:
              dir: examples/generic
              os: [ "darwin","linux" ]
              arch: [ "amd64" ]
              upx:
                enable: false
                params: [ "-min" ]
            004:
              dir: examples/demo
          upx:
            enable: true
            params: [ "-best", "-clean" ]

`
	// var b []byte
	var err error

	var holder Doc
	err = yaml.Unmarshal([]byte(src), &holder)
	t.Logf("holder: %+v", holder)

	var projectGroup = holder.Projects["000-default-group"]
	var p1 = projectGroup.Items["003-generic"]
	var p2 = projectGroup.Items["004"]
	var c = evendeep.New(
		// evendeep.WithStrategies(cms.OmitIfZero),
		evendeep.WithCopyStrategyOpt,
	)

	t.Run("copy upx", func(t *testing.T) {
		var upx Upx
		if err = c.CopyTo(holder.Upx, &upx); err != nil {
			t.Fatal(err)
		}
		if err = c.CopyTo(projectGroup.Upx, &upx); err != nil {
			t.Fatal(err)
		} else {
			if upx.Enable != true {
				t.Fatal("expecting true")
			}
			if !evendeep.DeepEqual(upx.Params, []string{"-best", "-clean"}) {
				t.Fatalf("expecting %v but got %v", []string{"-best", "-clean"}, upx.Params)
			}
		}
	})

	t.Run("copy os-arch", func(t *testing.T) {
		var common Common
		if err = c.CopyTo(holder.Common, &common); err != nil {
			t.Fatal(err)
		}
		if err = c.CopyTo(p1.Common, &common); err != nil {
			t.Fatal(err)
		} else {
			if !evendeep.DeepEqual(common.Os, []string{"darwin", "linux"}) {
				t.Fatalf("expecting %v but got %v", []string{"darwin", "linux"}, common.Os)
			}
			if !evendeep.DeepEqual(common.Arch, []string{"amd64"}) {
				t.Fatalf("expecting %v but got %v", []string{"amd64"}, common.Arch)
			}
		}

		if err = c.CopyTo(holder.Common, &common); err != nil {
			t.Fatal(err)
		}
		if err = c.CopyTo(p2.Common, &common); err != nil {
			t.Fatal(err)
		} else {
			if !evendeep.DeepEqual(common.Os, []string{"linux"}) {
				t.Fatalf("expecting %v but got %v", []string{"linux"}, common.Os)
			}
			if !evendeep.DeepEqual(common.Arch, []string{"amd64", "arm64"}) {
				t.Fatalf("expecting %v but got %v", []string{"amd64", "arm64"}, common.Arch)
			}
		}
	})
}

func Rindirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func TestChan01(t *testing.T) {
	t.Run(`check cap`, func(t *testing.T) {
		H := make(chan int, 5)
		v := reflect.ValueOf(H)
		t.Logf(`H.cap: %v`, v.Cap())

		H = make(chan int)
		v = reflect.ValueOf(H)
		t.Logf(`H.cap: %v`, v.Cap())
	})

	t.Run(`set field`, func(t *testing.T) {
		type A struct {
			H chan int
		}
		var a A
		v := reflect.ValueOf(&a)
		vind := Rindirect(v)
		fv := vind.FieldByName(`H`)
		cv := reflect.MakeChan(fv.Type(), 5)
		fv.Set(cv)
		t.Logf(`a: %v`, a)
	})
}

func TestComplexStructure901(t *testing.T) {
	const src = `---
app:
  bgo:
    build:
      for:
        - "linux/amd64"
        - "windows/amd64"
        - "darwin/amd64"
        - "darwin/arm64"
      os: [ linux ]
      arch: [ amd64, arm64 ]
      reduce: true
      upx:
        enable: false
        params: [ "-9" ]

      projects:                       # Projects map[string]ProjectGroup
        000-default-group:
          leading-text: dummy-title
          items:                      # Items    map[string]*ProjectWrap
            003-generic:
              dir: examples/generic
              os: [ "darwin","linux" ]
              arch: [ "amd64" ]
              upx:
                enable: false
                params: [ "-min" ]
            # 004:
            #   dir: examples/demo
          upx:
            enable: true
            params: [ "-best", "-clean" ]

`
	// var b []byte
	var err error

	var holder Doc
	err = yaml.Unmarshal([]byte(src), &holder)
	t.Logf("holder: %+v", holder)

	var c = evendeep.New(
		// evendeep.WithStrategies(cms.OmitIfZero),
		// evendeep.WithCopyStrategyOpt,
	)

	// t.Run("copy project-group map", func(t *testing.T) {
	// 	var conf Doc
	// 	// var projectGroup = holder.Projects["000-default-group"]
	// 	if err = c.CopyTo(holder.Projects, &conf.Projects); err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	if !evendeep.DeepEqual(holder, &conf) {
	// 		t.Fatalf("expecting \n\n%+v\n\nbut got\n\n%+v", holder, conf)
	// 	}
	// })

	t.Run("copy all Doc", func(t *testing.T) {
		var conf Doc
		if err = c.CopyTo(holder, &conf); err != nil {
			t.Fatalf("%+v", err)
		}
		if diff, equal := evendeep.DeepDiff(holder, conf); !equal {
			t.Fatalf("expecting \n\n%+v\n\nbut got\n\n%+v\n\ndiff is:\n\n%+v", holder, conf, diff)
		}
	})
}

// -----------------------------------------------------

type Doc struct {
	App `yaml:"app,omitempty"`
}

type App struct {
	Build `yaml:"bgo,omitempty"`
}

type Build struct {
	BgoSettings `yaml:"build,omitempty"`
}

// -----------------------------------------------------

type (
	//nolint:lll //no
	CommonBase struct {
		OS   string `yaml:"-" json:"-" toml:"-"` // just for string template expansion
		ARCH string `yaml:"-" json:"-" toml:"-"` // just for string template expansion

		Args       []string `yaml:"args,omitempty,flow" json:"args,omitempty" toml:"args,omitempty"`
		Ldflags    []string `yaml:"ldflags,omitempty,flow" json:"ldflags,omitempty" toml:"ldflags,omitempty"`          // default ldflags is to get the smaller build for releasing
		Asmflags   []string `yaml:"asmflags,omitempty,flow" json:"asmflags,omitempty" toml:"asmflags,omitempty"`       //
		Gcflags    []string `yaml:"gcflags,omitempty,flow" json:"gcflags,omitempty" toml:"gcflags,omitempty"`          //
		Gccgoflags []string `yaml:"gccgoflags,omitempty,flow" json:"gccgoflags,omitempty" toml:"gccgoflags,omitempty"` //
		Tags       []string `yaml:"tags,omitempty,flow" json:"tags,omitempty" toml:"tags,omitempty"`                   //
		// Cgo option
		Cgo bool `yaml:",omitempty" json:"cgo,omitempty" toml:"cgo,omitempty"` //
		// Race option enables data race detection.
		//		Supported only on linux/amd64, freebsd/amd64, darwin/amd64, windows/amd64,
		//		linux/ppc64le and linux/arm64 (only for 48-bit VMA).
		Race bool `yaml:",omitempty" json:"race,omitempty" toml:"race,omitempty"` //
		// Msan option enables interoperation with memory sanitizer.
		//		Supported only on linux/amd64, linux/arm64
		//		and only with Clang/LLVM as the host C compiler.
		//		On linux/arm64, pie build mode will be used.
		Msan bool `yaml:",omitempty" json:"msan,omitempty" toml:"msan,omitempty"` //
		// Asan enables interoperation with address sanitizer.
		//		Supported only on linux/arm64, linux/amd64.
		//		Supported only on linux/amd64 or linux/arm64 and only with GCC 7 and higher
		//		or Clang/LLVM 9 and higher.
		Asan bool `yaml:",omitempty" json:"asan,omitempty" toml:"asan,omitempty"` //
		//
		Mod           string `yaml:",omitempty" json:"mod,omitempty" toml:"mod,omitempty"`                                     // -mod reaonly,vendor,mod
		Amd64         string `yaml:",omitempty" json:"goamd64,omitempty" toml:"goamd64,omitempty"`                             // GOAMD64 v1,v2,v3,v4
		Gocmd         string `yaml:",omitempty" json:"gocmd,omitempty" toml:"gocmd,omitempty"`                                 // -gocmd go
		Gen           bool   `yaml:",omitempty" json:"gen,omitempty" toml:"gen,omitempty"`                                     // go generate at first?
		Install       bool   `yaml:",omitempty" json:"install,omitempty" toml:"install,omitempty"`                             // install binary to $GOPATH/bin like 'go install' ?
		Debug         bool   `yaml:",omitempty" json:"debug,omitempty" toml:"debug,omitempty"`                                 // true to produce a larger build with debug info
		DisableResult bool   `yaml:"disable-result,omitempty" json:"disable_result,omitempty" toml:"disable_result,omitempty"` // no ll (Shell list) building result

		// -X for -ldflags,
		// -X importpath.name=value
		//    Set the value of the string variable in importpath named name to value.
		//    Note that before Go 1.5 this option took two separate arguments.
		//    Now it takes one argument split on the first = sign.
		Extends      []PackageNameValues `yaml:"extends,omitempty" json:"extends,omitempty" toml:"extends,omitempty"` //
		CmdrSpecials bool                `yaml:"cmdr,omitempty" json:"cmdr,omitempty" toml:"cmdr,omitempty"`
		CmdrVersion  string              `yaml:"cmdr-version,omitempty" json:"cmdr-version,omitempty" toml:"cmdr-version,omitempty"`
	}

	// Upx params package.
	Upx struct {
		Enable bool     `yaml:",omitempty" json:"enable,omitempty" toml:"enable,omitempty"`
		Params []string `yaml:",omitempty" json:"params,omitempty" toml:"params,omitempty"`
	}

	PackageNameValues struct {
		Package string            `yaml:"pkg,omitempty" json:"package,omitempty" toml:"package,omitempty"`
		Values  map[string]string `yaml:"values,omitempty" json:"values,omitempty" toml:"values,omitempty"`
	}

	//nolint:lll //no
	Common struct {
		CommonBase `yaml:",omitempty,inline,flow" json:",omitempty" toml:""`

		Disabled       bool     `yaml:"disabled,omitempty" json:"disabled,omitempty" toml:"disabled,omitempty"`
		KeepWorkdir    bool     `yaml:"keep-workdir,omitempty" json:"keep_workdir,omitempty" toml:"keep_workdir,omitempty"`
		UseWorkDir     string   `yaml:"use-workdir,omitempty" json:"use_work_dir,omitempty" toml:"use_work_dir,omitempty"`
		For            []string `yaml:"for,omitempty,flow" json:"for,omitempty" toml:"for,omitempty"`
		Os             []string `yaml:"os,omitempty,flow" json:"os,omitempty" toml:"os,omitempty"`
		Arch           []string `yaml:"arch,omitempty,flow" json:"arch,omitempty" toml:"arch,omitempty"`
		Goroot         string   `yaml:"goroot,omitempty,flow" json:"goroot,omitempty" toml:"goroot,omitempty"`
		PreAction      string   `yaml:"pre-action,omitempty" json:"pre_action,omitempty" toml:"pre_action,omitempty"`                   // bash script
		PostAction     string   `yaml:"post-action,omitempty" json:"post_action,omitempty" toml:"post_action,omitempty"`                // bash script
		PreActionFile  string   `yaml:"pre-action-file,omitempty" json:"pre_action_file,omitempty" toml:"pre_action_file,omitempty"`    // bash script
		PostActionFile string   `yaml:"post-action-file,omitempty" json:"post_action_file,omitempty" toml:"post_action_file,omitempty"` // bash script

		DeepReduce bool `yaml:"reduce,omitempty" json:"reduce,omitempty" toml:"reduce,omitempty"`
		Upx        Upx  `yaml:",omitempty" json:"upx,omitempty" toml:"upx,omitempty"`
	}
)

type (
	DynBuildInfo struct {
		ProjectName         string
		AppName             string
		Version             string // copied from Project.Version
		BgoGroupKey         string // project-group key in .bgo.yml
		BgoGroupLeadingText string // same above,
		HasGoMod            bool   //
		GoModFile           string //
		GOROOT              string // force using a special GOROOT
		Dir                 string
	}

	Info struct {
		GoVersion      string // the result from 'go version'
		GoVersionMajor int    // 1
		GoVersionMinor int    // 17
		GitVersion     string // the result from `git describe --tags --abbrev=0`
		GitRevision    string // revision, git hash code, from `git rev-parse --short HEAD`
		GitSummary     string // `git describe --tags --dirty --always`
		GitDesc        string // `git log --online -1`
		BuilderComment string //
		BuildTime      string // format is like '2023-01-22T09:26:07+08:00'
		GOOS           string // a copy from runtime.GOOS
		GOARCH         string // a copy from runtime.GOARCH
		GOVERSION      string // a copy from runtime.Version()

		RandomString string
		RandomInt    int
		Serial       int64
	}
)

type TargetPlatforms struct {
	// OsArchMap is a map by key OS and CPUArch
	// key1: os
	// key2: arch
	OsArchMap map[string]map[string]bool `yaml:"os-arch-map"`
	// Sources is a map by key PackageName
	// key: packageName
	Sources map[string]*DynBuildInfo `yaml:"sources"`
}

type (
	ProjectWrap struct {
		Project `yaml:",omitempty,inline,flow" json:"project" toml:"project,omitempty"`
	}
	//nolint:lll //no
	Project struct {
		Name    string `yaml:"name,omitempty" json:"name,omitempty" toml:"name,omitempty"`          // appName if specified
		Dir     string `yaml:"dir" json:"dir" toml:"dir"`                                           // root dir of this module/cli-app
		Package string `yaml:"package,omitempty" json:"package,omitempty" toml:"package,omitempty"` // pkgName or mainGo
		MainGo  string `yaml:"main-go,omitempty" json:"main_go,omitempty" toml:"main_go,omitempty"` // default: 'main.go', // or: "./cli", "./cli/main.go", ...
		Version string `yaml:"version,omitempty" json:"version,omitempty" toml:"version,omitempty"` //

		// In current and earlier releases, MainGo is reserved and never used.

		*Common `yaml:",omitempty,inline,flow" json:",omitempty" toml:",omitempty"`

		// Cmdr building opts

		// More Posts:
		//   packaging methods
		//   coverage
		//   uploading
		//   docker
		//   github release
		// Before
		//   fmt, lint, test, cyclo, ...

		keyName string
		tp      *TargetPlatforms `copy:"-"`
	}
)

type (
	// BgoSettings for go build.
	// Generally it's loaded from '.bgo.yml'.
	BgoSettings struct {
		*Common `yaml:",omitempty,inline" json:",omitempty" toml:",omitempty"`

		Scope    string                  `yaml:"scope,omitempty" json:"scope,omitempty" toml:"scope,omitempty"`
		Projects map[string]ProjectGroup `yaml:"projects" json:"projects" toml:"projects"`

		Output   Output   `yaml:",omitempty" json:"output" toml:"output,omitempty"`
		Excludes []string `yaml:",omitempty,flow" json:"excludes,omitempty" toml:"excludes,omitempty"`
		Goproxy  string   `yaml:",omitempty" json:"goproxy,omitempty" toml:"goproxy,omitempty"`

		SavedAs []string `yaml:"saved-as,omitempty" json:"saved_as,omitempty" toml:"saved_as,omitempty"`
	}

	//nolint:lll //no
	Output struct {
		Dir         string `yaml:",omitempty" json:"dir,omitempty" toml:"dir,omitempty"`                   // Base Dir
		SplitTo     string `yaml:"split-to,omitempty" json:"split_to,omitempty" toml:"split_to,omitempty"` // see build.Context, Group, Project, ... | see also BuildContext struct
		NamedAs     string `yaml:"named-as,omitempty" json:"named_as,omitempty" toml:"named_as,omitempty"` // see build.Context,
		SuffixAs    string `yaml:"suffix-as,omitempty" json:"-" toml:"-"`                                  // NEVER USED.
		ZipSuffixAs string `yaml:"zip-suffix-as,omitempty" json:"-" toml:"-"`                              // NEVER USED. supported: gz/tgz, bz2, xz/txz, 7z
	}

	ProjectGroup struct {
		LeadingText string                  `yaml:"leading-text,omitempty" json:"leading_text,omitempty" toml:"leading_text,omitempty"` //nolint:lll //no
		Items       map[string]*ProjectWrap `yaml:"items" json:"items,omitempty" toml:"items,omitempty"`
		*Common     `yaml:",omitempty,inline" json:",omitempty" toml:",omitempty"`
	}
)
