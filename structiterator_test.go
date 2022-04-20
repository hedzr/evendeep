package evendeep

import (
	"fmt"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/typ"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/dbglog"
	"github.com/hedzr/evendeep/internal/tool"
)

func TestTableRecT_FieldName(t *testing.T) {
	tr := tableRecT{
		names: nil,
	}
	if tr.FieldName() != "" {
		t.FailNow()
	}

	tr = tableRecT{
		names: []string{"C", "B", "A"},
	}
	if tr.FieldName() != "A.B.C" {
		t.FailNow()
	}
}

func TestTableRecT_ShortFieldName(t *testing.T) {
	tr := tableRecT{
		names: nil,
	}
	if tr.ShortFieldName() != "" {
		t.FailNow()
	}
}

func TestSetToZero(t *testing.T) {
	val := 10
	type Tmp struct {
		A int16
	}
	fn := func() {}

	run := func(v reflect.Value) {
		// v := reflect.ValueOf(c)
		vind := tool.Rdecodesimple(v)
		vf := vind.Field(0)
		// vf = vf.Elem()
		t.Logf("src: %v", tool.Valfmt(&vf))
		setToZero(&vf)
	}

	t.Run("TestSetToZero.bool", func(t *testing.T) {
		type Case struct {
			Src, Zero bool
		}
		c := &Case{true, false}
		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.int", func(t *testing.T) {
		type Case struct {
			Src, Zero int
		}
		c := &Case{9, 0}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.uint", func(t *testing.T) {
		type Case struct {
			Src, Zero uint
		}
		c := &Case{9, 0}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.float", func(t *testing.T) {
		type Case struct {
			Src, Zero float32
		}
		c := &Case{9, 0}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.complex", func(t *testing.T) {
		type Case struct {
			Src, Zero complex128
		}
		c := &Case{9, 0}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.string", func(t *testing.T) {
		type Case struct {
			Src, Zero string
		}
		c := &Case{"9hello", ""}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.slice", func(t *testing.T) {
		type Case struct {
			Src, Zero []int
		}
		c := &Case{[]int{9, 12}, []int{}}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.array", func(t *testing.T) {
		type Case struct {
			Src, Zero [1]int
		}
		c := &Case{[1]int{1}, [1]int{0}}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.map", func(t *testing.T) {
		type Case struct {
			Src, Zero map[string]int
		}
		c := &Case{map[string]int{"x": 1, "y": 2}, map[string]int{}}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.ptr", func(t *testing.T) {
		type Case struct {
			Src, Zero *int
		}
		c := &Case{&val, (*int)(nil)}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})
	t.Run("TestSetToZero.struct", func(t *testing.T) {
		type Case struct {
			Src, Zero Tmp
		}
		c := &Case{Tmp{A: 3}, Tmp{}}

		v := reflect.ValueOf(c)
		run(v)
		if !DeepEqual(c.Src, c.Zero) {
			t.Logf("-. For source '%v', expect '%v' but failed.", c.Src, c.Zero)
			t.FailNow()
		}
	})

	for i, case_ := range []*struct {
		Src  typ.Any
		Zero typ.Any
	}{
		{true, false},
		{9, 0},
		{uint32(9), uint32(0)},
		{float64(9), float64(0)},
		{complex128(9), complex128(0)},
		{"9hello", ""},
		{[]int{9, 12}, []int{}},
		{[1]int{1}, [1]int{0}},
		{map[string]int{"x": 1, "y": 2}, map[string]int{}},
		{&val, (*int)(nil)},

		{Tmp{A: 3}, Tmp{}},

		{fn, fn},
	} {
		v := reflect.ValueOf(case_)
		vind := tool.Rdecodesimple(v)
		vf := vind.Field(0)
		// vf = vf.Elem()
		t.Logf("src: %v", tool.Valfmt(&vf))
		setToZero(&vf)
		if !DeepEqual(case_.Src, case_.Zero) {
			t.Logf("%d. For source '%v', expect '%v' but failed.", i, case_.Src, case_.Zero)
			t.FailNow()
		}
	}

	// for unsupported types

	v := reflect.ValueOf(&fn).Elem()
	setToZero(&v)
}

func TestIgnoredPackagePrefixesContains(t *testing.T) {
	_ignoredpackageprefixes.contains("abc")
	_ignoredpackageprefixes.contains("golang.org/grpc")
}

func TestFieldAccessorT_IsAnyFlagsOK(t *testing.T) {
	const byName = false
	x2 := x2data()
	v1 := reflect.ValueOf(&x2)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1)
	loopIt(t, it, byName, nil)
}

func TestFieldAccessorT_FieldType_is_nil(t *testing.T) {
	var a *fieldAccessorT = nil
	_ = a.FieldType()

	a = &fieldAccessorT{isStruct: false}
	_ = a.NumField()
}

type Part struct {
	S string
}

func (s Part) String() string {
	return s.S
}

func TestFieldAccessorT_forMap(t *testing.T) {
	const byName = false
	x2 := map[Part]bool{
		Part{"x"}: true, //nolint:gofmt
	}
	v1 := reflect.ValueOf(&x2)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1)
	loopIt(t, it, byName, nil)
}

func TestFieldAccessorT_Operations(t *testing.T) {
	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := tool.Rdecode(v1)

	// x2 := new(X1)
	// t2 := rdecodesimple(reflect.ValueOf(&x2))

	const byName = false

	it := newStructIterator(t1)
	loopIt(t, it, byName, func(accessor accessor, field *reflect.StructField) {
		if field.Name == "O" {
			accessor.Set(reflect.ValueOf([2]string{"zed", "bee"}))
		}
	})

	// for i := 0; ; i++ {
	// 	accessor, ok := it.Next(nil, byName)
	// 	if !ok {
	// 		break
	// 	}
	// 	field := accessor.StructField()
	// 	if field == nil {
	// 		t.Logf("%d. field info missed", i)
	// 		continue
	// 	}
	// 	if field.Name == "O" {
	// 		accessor.Set(reflect.ValueOf([2]string{"zed", "bee"}))
	// 	}
	// 	t.Logf("%d. %q (%v) %v %q | %v", i, field.Name, tool.Typfmt(field.Type), field.Index, field.PkgPath, accessor.FieldValue().Interface())
	// }
}

func TestStructIteratorT_Next_X1(t *testing.T) {

	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := tool.Rdecode(v1)

	x2 := new(X1)
	t2 := tool.Rdecodesimple(reflect.ValueOf(&x2))

	const byName = false

	t.Run("getAllFields at once", testGetAllFieldsX1)

	t.Run("by struct iterator", testStructIteratorNextT1)

	t.Run("get further: loop src & dst at same time", func(t *testing.T) {

		targetIterator := newStructIterator(t1)

		var sourcefields fieldsTableT
		sourcefields = sourcefields.getAllFields(t1, false)
		for i, amount := 0, len(sourcefields.tableRecordsT); i < amount; i++ {
			sourcefield := sourcefields.tableRecordsT[i]
			flags := parseFieldTags(sourcefield.structField.Tag, "")
			accessor, ok := targetIterator.Next(nil, byName) //nolint:govet
			if flags.isFlagExists(cms.Ignore) || !ok {
				continue
			}
			srcval, dstval := sourcefield.FieldValue(), accessor.FieldValue()
			if srcval == nil || dstval == nil {
				t.Logf("%d. field info missed", i)
				continue
			}
			t.Logf("%d. %s (%v) -> %s (%v) | %v", i, strings.Join(tool.ReverseStringSlice(sourcefield.names), "."), tool.Valfmt(srcval), accessor.StructFieldName(), tool.Valfmt(dstval), tool.Typfmt(accessor.StructField().Type))
			// ec.Attach(invokeStructFieldTransformer(c, params, srcval, dstval, padding))
		}

	})

	t.Run("zero: loop src & dst at same time", func(t *testing.T) {

		targetIterator := newStructIterator(t2, withStructPtrAutoExpand(true))

		var sourcefields fieldsTableT
		sourcefields = sourcefields.getAllFields(t2, true)

		for i, amount := 0, len(sourcefields.tableRecordsT); i < amount; i++ {
			sourcefield := sourcefields.tableRecordsT[i]
			flags := parseFieldTags(sourcefield.structField.Tag, "")
			accessor, ok := targetIterator.Next(nil, byName) //nolint:govet
			if flags.isFlagExists(cms.Ignore) || !ok {
				continue
			}
			srcval, dstval := sourcefield.FieldValue(), accessor.FieldValue()
			if srcval == nil || dstval == nil {
				t.Logf("%d. field info missed", i)
				continue
			}
			t.Logf("%d. %s (%v) -> %s (%v) | %v", i, strings.Join(tool.ReverseStringSlice(sourcefield.names), "."), tool.Valfmt(srcval), accessor.StructFieldName(), tool.Valfmt(dstval), tool.Typfmt(accessor.StructField().Type))
			// ec.Attach(invokeStructFieldTransformer(c, params, srcval, dstval, padding))
		}

	})

}

func loopIt(t *testing.T, it structIterable, byName bool, checkName func(accessor accessor, field *reflect.StructField)) {
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		if accessor.IsStruct() && field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}

		if checkName != nil {
			checkName(accessor, field)
		}

		var name string
		var fieldType *reflect.Type
		if accessor.IsStruct() {
			name = field.Name
			fieldType = &field.Type
		} else {
			name = accessor.StructFieldName()
			fieldType = accessor.FieldType()
		}

		// dbglog.Log("  - for field %q", name)

		if name == "Name" {
			accessor.Set(reflect.ValueOf("Meg"))
			if p, ok := accessor.(*fieldAccessorT); ok {
				p.fieldTags = parseFieldTags(field.Tag, "")
				accessor.IsAnyFlagsOK(cms.Default, cms.KeepIfNotEq)
				accessor.IsAllFlagsOK(cms.Default, cms.KeepIfNotEq)
			}
		} else {
			if p, ok := accessor.(*fieldAccessorT); ok {
				p.fieldTags = nil
				accessor.IsAnyFlagsOK(cms.Default, cms.KeepIfNotEq)
				accessor.IsAllFlagsOK(cms.Default, cms.KeepIfNotEq)
			}
		}

		var val reflect.Value
		if accessor.IsStruct() && !tool.IsExported(field) {
			val = *accessor.FieldValue()
			val = cl.GetUnexportedField(val)
			// cl.SetUnexportedField(target, newval)
		} else {
			val = *accessor.FieldValue()
		}
		t.Logf("%d. %q (%v) | %v", i, name, tool.Typfmt(*fieldType), tool.Valfmt(&val))
	}
}

func x1data() X1 {
	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"

	x0 := X0{}

	x1 := X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:3],
		O: a,
		Q: a,
	}
	return x1
}

func x2data() Employee {
	return Employee{
		Name: "Bob",
	}
}

func testGetAllFieldsX1(t *testing.T) {

	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := tool.Rdecode(v1)

	var sourcefields fieldsTableT
	sourcefields = sourcefields.getAllFields(t1, false)
	for i, amount := 0, len(sourcefields.tableRecordsT); i < amount; i++ {
		sourcefield := sourcefields.tableRecordsT[i]
		srcval := sourcefield.FieldValue()
		srctypstr := tool.Typfmtv(srcval)
		dbglog.Log("%d. %s (%v) %v -> %s (%v)", i, strings.Join(tool.ReverseStringSlice(sourcefield.names), "."), tool.Valfmt(srcval), srctypstr, "", "")
	}
}

func testStructIteratorNextT1(t *testing.T) {

	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := tool.Rdecode(v1)

	const byName = false

	it := newStructIterator(t1)
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		dbglog.Log("%d. %q (%v) %v %q", i, field.Name, tool.Typfmt(field.Type), field.Index, field.PkgPath)
	}
}

func TestStructIterator_Next_Employee2(t *testing.T) {
	t.Run("testStructIterator_Next_Employee2", testStructIteratorNextEmployee2)
	t.Run("testStructIterator_Next_Employee2_exp", testStructIteratorNextEmployee2Exp)
}

func testStructIteratorNextEmployee2(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)

	src := Employee2{
		Base: Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &Attr{Attrs: []string{"hello", "world"}},
		Valid:  true,
	}

	var sb strings.Builder
	defer func() { t.Log(sb.String()) }()

	v1 := reflect.ValueOf(&src)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1)
	const byName = false
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		t.Logf("%d. %q (%v) %v %q", i, field.Name, tool.Typfmt(field.Type), field.Index, field.PkgPath)
		sb.WriteString(fmt.Sprintf("%d. %q (%v) %v %q\n", i, field.Name, tool.Typfmt(field.Type), field.Index, field.PkgPath))
	}

	if sb.String() != `0. "Base" (evendeep.Base (struct)) [0] ""
1. "Avatar" (string (string)) [1] ""
2. "Image" ([]uint8 (slice)) [2] ""
3. "Attr" (*evendeep.Attr (ptr)) [3] ""
4. "Valid" (bool (bool)) [4] ""
5. "Deleted" (bool (bool)) [5] ""
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}
}

func testStructIteratorNextEmployee2Exp(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)

	src := Employee2{
		Base: Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &Attr{Attrs: []string{"hello", "world"}},
		Valid:  true,
	}

	var sb strings.Builder
	defer func() { t.Log(sb.String()) }()

	v1 := reflect.ValueOf(&src)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true))
	const byName = false
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		t.Logf("%d. %q (%v) %v %q", i, field.Name, tool.Typfmt(field.Type), field.Index, field.PkgPath)
		sb.WriteString(fmt.Sprintf("%d. %q (%v) %v %q\n", i, field.Name, tool.Typfmt(field.Type), field.Index, field.PkgPath))
	}

	if sb.String() != `0. "Name" (string (string)) [0] ""
1. "Birthday" (*time.Time (ptr)) [1] ""
2. "Age" (int (int)) [2] ""
3. "EmployeID" (int64 (int64)) [3] ""
4. "Avatar" (string (string)) [1] ""
5. "Image" ([]uint8 (slice)) [2] ""
6. "Attrs" ([]string (slice)) [0] ""
7. "Valid" (bool (bool)) [4] ""
8. "Deleted" (bool (bool)) [5] ""
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}
}

func TestStructIterator_Next_User(t *testing.T) {
	// new target with new children automatically
	t.Run("testStructIterator_Next_User_new", testStructIteratorNextUserNew)

	t.Run("testStructIterator_Next_User", testStructIteratorNextUser)
	t.Run("testStructIterator_Next_User_zero", testStructIteratorNextUserZero)
	t.Run("testStructIterator_Next_User_more", testStructIteratorNextUserMore)
}

func testStructIteratorNextUser(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	tgt := User{
		Name:      "Frank",
		Birthday:  &tm2,
		Age:       18,
		EmployeID: 9,
		Attr:      &Attr{Attrs: []string{"baby"}},
		Deleted:   true,
	}

	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	v1 := reflect.ValueOf(&tgt)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1)
	const byName = false
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v) | fld: %+v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index, field)
	}

}

func testStructIteratorNextUserNew(t *testing.T) {

	// timeZone, _ := time.LoadLocation("America/Phoenix")
	// timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// //tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// //tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	// tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	const byName = false
	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, tgt := range []*User{
		new(User),
	} {
		sb.WriteString("\n")
		v1 := reflect.ValueOf(&tgt) // nolint:gosec // G601: Implicit memory aliasing in for loop
		t1, _ := tool.Rdecode(v1)
		it := newStructIterator(t1, withStructPtrAutoExpand(true), withStructFieldPtrAutoNew(true))
		for i := 0; ; i++ {
			accessor, ok := it.Next(nil, false)
			if !ok {
				break
			}
			field := accessor.StructField()
			_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v) | fld: %+v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index, field)
		}
	}

}

func testStructIteratorNextUserZero(t *testing.T) {

	// timeZone, _ := time.LoadLocation("America/Phoenix")
	// timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	// tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	const byName = false
	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, tgt := range []*User{
		new(User),
	} {
		sb.WriteString("\n\n\n")
		v1 := reflect.ValueOf(&tgt) // nolint:gosec // G601: Implicit memory aliasing in for loop
		t1, _ := tool.Rdecode(v1)
		it := newStructIterator(t1)
		for i := 0; ; i++ {
			accessor, ok := it.Next(nil, byName)
			if !ok {
				break
			}
			field := accessor.StructField()
			_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v) | fld: %+v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index, field)
		}
	}

}

func testStructIteratorNextUserMore(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	const byName = false
	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, tgt := range []*User{
		{
			Name:      "Frank",
			Birthday:  &tm2,
			Age:       18,
			EmployeID: 9,
			Attr:      &Attr{Attrs: []string{"baby"}},
		},
		{
			Name:      "Mathews",
			Birthday:  &tm3,
			Age:       3,
			EmployeID: 92,
			Attr:      &Attr{Attrs: []string{"get"}},
			Deleted:   false,
		},
	} {
		sb.WriteString("\n\n\n")
		v1 := reflect.ValueOf(&tgt) // nolint:gosec // G601: Implicit memory aliasing in for loop
		t1, _ := tool.Rdecode(v1)
		it := newStructIterator(t1)
		for i := 0; ; i++ {
			accessor, ok := it.Next(nil, byName)
			if !ok {
				break
			}
			field := accessor.StructField()
			_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v) | fld: %+v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index, field)
		}
	}

}

func TestStructIterator_Next_A4(t *testing.T) {
	t.Run("testStructIterator_Next_A4_new", teststructiteratorNextA4New)

	t.Run("testStructIterator_Next_A4_zero", teststructiteratorNextA4Zero)
	t.Run("testStructIterator_Next_A4", teststructiteratorNextA4)
}

func teststructiteratorNextA4New(t *testing.T) {
	var sb strings.Builder
	const byName = false
	a4 := new(A4)
	v1 := reflect.ValueOf(&a4)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true), withStructFieldPtrAutoNew(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v %v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index)
	}

	t.Logf(sb.String())
	//nolint:goconst
	if sb.String() != `0. "Name2" (string (string)) | evendeep.A2 (struct) [0]
1. "Int2" (int (int)) | evendeep.A2 (struct) [1]
2. "Bool2" (bool (bool)) | evendeep.A2 (struct) [2]
3. "Name1" (string (string)) | evendeep.A1 (struct) [0]
4. "Int1" (int (int)) | evendeep.A1 (struct) [1]
5. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
6. "Name3" (string (string)) | evendeep.A3 (struct) [1]
7. "Int3" (int (int)) | evendeep.A3 (struct) [2]
8. "Name1" (string (string)) | evendeep.A1 (struct) [0]
9. "Int1" (int (int)) | evendeep.A1 (struct) [1]
10. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
11. "Bool3" (bool (bool)) | evendeep.A3 (struct) [4]
12. "Int4" (int (int)) | evendeep.A4 (struct) [1]
13. "Name1" (string (string)) | evendeep.A1 (struct) [0]
14. "Int1" (int (int)) | evendeep.A1 (struct) [1]
15. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. "Name2" (string (string)) | evendeep.A2 (struct) (0) [0]
	// 1. "Int2" (int (int)) | evendeep.A2 (struct) (1) [1]
	// 2. "Bool2" (bool (bool)) | evendeep.A2 (struct) (2) [2]
	// 3. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 4. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 5. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
	// 6. "Name3" (string (string)) | evendeep.A3 (struct) (1) [1]
	// 7. "Int3" (int (int)) | evendeep.A3 (struct) (2) [2]
	// 8. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 9. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 10. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
	// 11. "Bool3" (bool (bool)) | evendeep.A3 (struct) (4) [4]
	// 12. "Int4" (int (int)) | evendeep.A4 (struct) (1) [1]
	// 13. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 14. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 15. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
}

func teststructiteratorNextA4Zero(t *testing.T) {
	var sb strings.Builder
	const byName = false
	a4 := new(A4)
	v1 := reflect.ValueOf(&a4)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v %v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index)
	}

	t.Logf(sb.String())
	if sb.String() != `0. "Name2" (string (string)) | evendeep.A2 (struct) [0]
1. "Int2" (int (int)) | evendeep.A2 (struct) [1]
2. "Bool2" (bool (bool)) | evendeep.A2 (struct) [2]
3. "Name1" (string (string)) | evendeep.A1 (struct) [0]
4. "Int1" (int (int)) | evendeep.A1 (struct) [1]
5. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
6. "Name3" (string (string)) | evendeep.A3 (struct) [1]
7. "Int3" (int (int)) | evendeep.A3 (struct) [2]
8. "Name1" (string (string)) | evendeep.A1 (struct) [0]
9. "Int1" (int (int)) | evendeep.A1 (struct) [1]
10. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
11. "Bool3" (bool (bool)) | evendeep.A3 (struct) [4]
12. "Int4" (int (int)) | evendeep.A4 (struct) [1]
13. "Name1" (string (string)) | evendeep.A1 (struct) [0]
14. "Int1" (int (int)) | evendeep.A1 (struct) [1]
15. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. "Name2" (string (string)) | evendeep.A2 (struct) (0) [0]
	// 1. "Int2" (int (int)) | evendeep.A2 (struct) (1) [1]
	// 2. "Bool2" (bool (bool)) | evendeep.A2 (struct) (2) [2]
	// 3. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 4. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 5. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
	// 6. "Name3" (string (string)) | evendeep.A3 (struct) (1) [1]
	// 7. "Int3" (int (int)) | evendeep.A3 (struct) (2) [2]
	// 8. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 9. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 10. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
	// 11. "Bool3" (bool (bool)) | evendeep.A3 (struct) (4) [4]
	// 12. "Int4" (int (int)) | evendeep.A4 (struct) (1) [1]
	// 13. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 14. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 15. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
}

func teststructiteratorNextA4(t *testing.T) {
	var sb strings.Builder
	const byName = false
	a4 := prepareDataA4()
	v1 := reflect.ValueOf(&a4)
	t1, _ := tool.Rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next(nil, byName)
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v %v\n", i, field.Name, tool.Typfmt(field.Type), tool.Typfmt(accessor.Type()), field.Index)
	}

	t.Logf(sb.String())
	if sb.String() != `0. "Name2" (string (string)) | evendeep.A2 (struct) [0]
1. "Int2" (int (int)) | evendeep.A2 (struct) [1]
2. "Bool2" (bool (bool)) | evendeep.A2 (struct) [2]
3. "Name1" (string (string)) | evendeep.A1 (struct) [0]
4. "Int1" (int (int)) | evendeep.A1 (struct) [1]
5. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
6. "Name3" (string (string)) | evendeep.A3 (struct) [1]
7. "Int3" (int (int)) | evendeep.A3 (struct) [2]
8. "Name1" (string (string)) | evendeep.A1 (struct) [0]
9. "Int1" (int (int)) | evendeep.A1 (struct) [1]
10. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
11. "Bool3" (bool (bool)) | evendeep.A3 (struct) [4]
12. "Int4" (int (int)) | evendeep.A4 (struct) [1]
13. "Name1" (string (string)) | evendeep.A1 (struct) [0]
14. "Int1" (int (int)) | evendeep.A1 (struct) [1]
15. "Bool1" (bool (bool)) | evendeep.A1 (struct) [2]
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. "Name2" (string (string)) | evendeep.A2 (struct) (0) [0]
	// 1. "Int2" (int (int)) | evendeep.A2 (struct) (1) [1]
	// 2. "Bool2" (bool (bool)) | evendeep.A2 (struct) (2) [2]
	// 3. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 4. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 5. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
	// 6. "Name3" (string (string)) | evendeep.A3 (struct) (1) [1]
	// 7. "Int3" (int (int)) | evendeep.A3 (struct) (2) [2]
	// 8. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 9. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 10. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
	// 11. "Bool3" (bool (bool)) | evendeep.A3 (struct) (4) [4]
	// 12. "Int4" (int (int)) | evendeep.A4 (struct) (1) [1]
	// 13. "Name1" (string (string)) | evendeep.A1 (struct) (0) [0]
	// 14. "Int1" (int (int)) | evendeep.A1 (struct) (1) [1]
	// 15. "Bool1" (bool (bool)) | evendeep.A1 (struct) (2) [2]
}

func TestFieldsTable_getallfields(t *testing.T) {
	t.Run("test_getallfields X1 deep", testFieldsTableGetFieldsDeeply)
	t.Run("test_getallfields", testFieldsTableGetAllFields)
	t.Run("test_getallfields_Employee2", testFieldsTableGetAllFieldsEmployee2)
	t.Run("test_getallfields_Employee2_2", testFieldsTableGetAllFieldsEmployee22)
	t.Run("test_getallfields_User", testFieldsTableGetAllFieldsUser)
}

func testFieldsTableGetFieldsDeeply(t *testing.T) {

	x2 := new(X1)
	t2 := tool.Rdecodesimple(reflect.ValueOf(&x2))

	var sourcefields fieldsTableT
	sourcefields.getAllFields(t2, true)

	var sb strings.Builder
	defer func() { t.Log(sb.String()) }()

	for i, f := range sourcefields.tableRecordsT {
		_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
			f.FieldName(), f.indexes,
			tool.Typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
		)
	}

}

func testFieldsTableGetAllFields(t *testing.T) {
	a4 := prepareDataA4()
	v4 := reflect.ValueOf(&a4)

	var sourcefields fieldsTableT
	sourcefields.getAllFields(v4, true)

	var sb strings.Builder
	defer func() { t.Log(sb.String()) }()

	for i, f := range sourcefields.tableRecordsT {
		_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
			f.FieldName(), f.indexes,
			tool.Typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
		)
	}

	if sb.String() != `0. A2.Name2, [0 0] | string (string), "", ""
1. A2.Int2, [1 0] | int (int), "", ""
2. A2.Bool2, [2 0] | bool (bool), "", ""
3. A1.Name1, [0 3] | string (string), "", ""
4. A1.Int1, [1 3] | int (int), "", ""
5. A1.Bool1, [2 3] | bool (bool), "", ""
6. A3.Name3, [1 0] | string (string), "", ""
7. A3.Int3, [2 0] | int (int), "", ""
8. A1.Name1, [0 3] | string (string), "", ""
9. A1.Int1, [1 3] | int (int), "", ""
10. A1.Bool1, [2 3] | bool (bool), "", ""
11. A3.Bool3, [4 0] | bool (bool), "", ""
12. Int4, [1] | int (int), "", ""
13. A1.Name1, [0 2] | string (string), "", ""
14. A1.Int1, [1 2] | int (int), "", ""
15. A1.Bool1, [2 2] | bool (bool), "", ""
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. [Name2 A2], [0 0] | &{Name2  string  0 [0] false}
	// 1. [Int2 A2], [1 0] | &{Int2  int  16 [1] false}
	// 2. [Bool2 A2], [2 0] | &{Bool2  bool  24 [2] false}
	// 3. [Name1 A1], [0 3] | &{Name1  string  0 [0] false}
	// 4. [Int1 A1], [1 3] | &{Int1  int  16 [1] false}
	// 5. [Bool1 A1], [2 3] | &{Bool1  bool  24 [2] false}
	// 6. [Name3 A3], [1 0] | &{Name3  string  8 [1] false}
	// 7. [Int3 A3], [2 0] | &{Int3  int  24 [2] false}
	// 8. [Name1 A1], [0 3] | &{Name1  string  0 [0] false}
	// 9. [Int1 A1], [1 3] | &{Int1  int  16 [1] false}
	// 10. [Bool1 A1], [2 3] | &{Bool1  bool  24 [2] false}
	// 11. [Bool3 A3], [4 0] | &{Bool3  bool  64 [4] false}
	// 12. [Int4], [1] | &{Int4  int  8 [1] false}
	// 13. [Name1 A1], [0 2] | &{Name1  string  0 [0] false}
	// 14. [Int1 A1], [1 2] | &{Int1  int  16 [1] false}
	// 15. [Bool1 A1], [2 2] | &{Bool1  bool  24 [2] false}
}

func testFieldsTableGetAllFieldsEmployee2(t *testing.T) {

	// timeZone, _ := time.LoadLocation("America/Phoenix")
	// timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	// tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	// var sb strings.Builder
	// defer func() {
	// 	t.Logf("\n%v\n", sb.String())
	// }()

	for _, a4 := range []*Employee2{
		new(Employee2),
		// {
		//	Base: Base{
		//		Name:      "Bob",
		//		Birthday:  &tm,
		//		Age:       24,
		//		EmployeID: 7,
		//	},
		//	Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		//	Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		//	Attr:   &Attr{Attrs: []string{"hello", "world"}},
		//	Valid:  true,
		// },
	} {
		// sb.WriteString("\n")

		v4 := reflect.ValueOf(&a4) // nolint:gosec // G601: Implicit memory aliasing in for loop

		var sourcefields fieldsTableT
		sourcefields.getAllFields(v4, true)

		var sb1 strings.Builder
		// defer func() { t.Log(sb.String()) }()

		for i, f := range sourcefields.tableRecordsT {
			_, _ = fmt.Fprintf(&sb1, "%v. %v, %v | %v, %q, %q\n", i,
				f.FieldName(), f.indexes,
				tool.Typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
			)
		}

		if sb1.String() != `0. Base.Name, [0 0] | string (string), "", ""
1. Base.Birthday, [1 0] | *time.Time (ptr), "", ""
2. Base.Age, [2 0] | int (int), "", ""
3. Base.EmployeID, [3 0] | int64 (int64), "", ""
4. Avatar, [1] | string (string), "", ""
5. Image, [2] | []uint8 (slice), "", ""
6. Attr.Attrs, [0 3] | []string (slice), "copy:\",slicemerge\"", ""
7. Valid, [4] | bool (bool), "", ""
8. Deleted, [5] | bool (bool), "", ""
` {
			t.Fail()
		} else {
			t.Log("The loops and outputs verified")
		}

	}
}

func testFieldsTableGetAllFieldsEmployee22(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	// timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	// tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	var sb strings.Builder
	defer func() { t.Logf("\n%v\n", sb.String()) }()

	for _, a4 := range []*Employee2{
		// new(Employee2),
		{
			Base: Base{
				Name:      "Bob",
				Birthday:  &tm,
				Age:       24,
				EmployeID: 7,
			},
			Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
			Image:  []byte{95, 27, 43, 66, 0, 21, 210},
			Attr:   &Attr{Attrs: []string{"hello", "world"}},
			Valid:  true,
		},
	} {
		sb.WriteString("\n")

		v4 := reflect.ValueOf(&a4) // nolint:gosec // G601: Implicit memory aliasing in for loop

		var sourcefields fieldsTableT
		sourcefields.getAllFields(v4, true)

		for i, f := range sourcefields.tableRecordsT {
			_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
				f.FieldName(), f.indexes,
				tool.Typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
			)
		}

		if sb.String() != `
0. Base.Name, [0 0] | string (string), "", ""
1. Base.Birthday, [1 0] | *time.Time (ptr), "", ""
2. Base.Age, [2 0] | int (int), "", ""
3. Base.EmployeID, [3 0] | int64 (int64), "", ""
4. Avatar, [1] | string (string), "", ""
5. Image, [2] | []uint8 (slice), "", ""
6. Attr.Attrs, [0 3] | []string (slice), "copy:\",slicemerge\"", ""
7. Valid, [4] | bool (bool), "", ""
8. Deleted, [5] | bool (bool), "", ""
` {
			t.Fail()
		} else {
			t.Log("The loops and outputs verified")
		}
	}
}

func testFieldsTableGetAllFieldsUser(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, a4 := range []*User{
		new(User),
		{
			Name:      "Frank",
			Birthday:  &tm2,
			Age:       18,
			EmployeID: 9,
			Attr:      &Attr{Attrs: []string{"baby"}},
		},
		{
			Name:      "Mathews",
			Birthday:  &tm3,
			Age:       3,
			EmployeID: 92,
			Attr:      &Attr{Attrs: []string{"get"}},
			Deleted:   false,
		},
	} {
		// sb.WriteString("\n")

		v4 := reflect.ValueOf(&a4) // nolint:gosec // G601: Implicit memory aliasing in for loop

		var sourcefields fieldsTableT
		sourcefields.getAllFields(v4, true)

		for i, f := range sourcefields.tableRecordsT {
			_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
				f.FieldName(), f.indexes,
				tool.Typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
			)
		}

		t.Logf("%v\n\n", sb.String())
		if sb.String() != `0. Name, [0] | string (string), "", ""
1. Birthday, [1] | *time.Time (ptr), "", ""
2. Age, [2] | int (int), "", ""
3. EmployeID, [3] | int64 (int64), "", ""
4. Avatar, [4] | string (string), "", ""
5. Image, [5] | []uint8 (slice), "", ""
6. Attr.Attrs, [0 6] | []string (slice), "copy:\",slicemerge\"", ""
7. Valid, [7] | bool (bool), "", ""
8. Deleted, [8] | bool (bool), "", ""
` {
			t.Fail()
		} else {
			t.Log("The loops and outputs verified")
		}

		sb.Reset()
	}
}

func prepareDataA4() *A4 {

	a4 := &A4{
		A3: &A3{
			A2: &A2{
				Name2: "",
				Int2:  0,
				Bool2: false,
				A1: A1{
					Name1: "",
					Int1:  0,
					Bool1: false,
				},
			},
			Name3: "",
			Int3:  0,
			A1: A1{
				Name1: "",
				Int1:  0,
				Bool1: false,
			},
			Bool3: false,
		},
		Int4: 0,
		A1: &A1{
			Name1: "",
			Int1:  0,
			Bool1: false,
		},
	}
	return a4
}

type A1 struct {
	Name1 string
	Int1  int
	Bool1 bool
}
type A2 struct {
	Name2 string
	Int2  int
	Bool2 bool
	A1
}
type A3 struct {
	*A2
	Name3 string
	Int3  int
	A1
	Bool3 bool
}
type A4 struct {
	A3   *A3
	Int4 int
	*A1
}
