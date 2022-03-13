package deepcopy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestIgnoredpackageprefixesContains(t *testing.T) {
	_ignoredpackageprefixes.contains("abc")
	_ignoredpackageprefixes.contains("golang.org/grpc")
}

func TestFieldaccessorOperations(t *testing.T) {
	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := rdecode(v1)

	//x2 := new(X1)
	//t2 := rdecodesimple(reflect.ValueOf(&x2))

	it := newStructIterator(t1)
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		if field.Name == "O" {
			accessor.Set(reflect.ValueOf([2]string{"zed", "bee"}))
		}
		t.Logf("%d. %q (%v) %v %q | %v", i, field.Name, typfmt(field.Type), field.Index, field.PkgPath, accessor.FieldValue().Interface())
	}
}

func TestStructIterator_Next_X1(t *testing.T) {

	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := rdecode(v1)

	x2 := new(X1)
	t2 := rdecodesimple(reflect.ValueOf(&x2))

	t.Run("getallfields at once", testgetallfieldsX1)

	t.Run("by struct iterator", teststructiteratorNextT1)

	t.Run("get further: loop src & dst at same time", func(t *testing.T) {

		targetIterator := newStructIterator(t1)

		var sourcefields fieldstable
		sourcefields = sourcefields.getallfields(t1, false)
		for i, amount := 0, len(sourcefields.tablerecords); i < amount; i++ {
			sourcefield := sourcefields.tablerecords[i]
			flags := parseFieldTags(sourcefield.structField.Tag)
			accessor, ok := targetIterator.Next()
			if flags.isFlagExists(Ignore) || !ok {
				continue
			}
			srcval, dstval := sourcefield.FieldValue(), accessor.FieldValue()
			if srcval == nil || dstval == nil {
				t.Logf("%d. field info missed", i)
				continue
			}
			t.Logf("%d. %s (%v) -> %s (%v) | %v", i, strings.Join(reverseStringSlice(sourcefield.names), "."), valfmt(srcval), accessor.StructFieldName(), valfmt(dstval), typfmt(accessor.StructField().Type))
			// ec.Attach(invokeStructFieldTransformer(c, params, srcval, dstval, padding))
		}

	})

	t.Run("zero: loop src & dst at same time", func(t *testing.T) {

		targetIterator := newStructIterator(t2, withStructPtrAutoExpand(true))

		var sourcefields fieldstable
		sourcefields = sourcefields.getallfields(t2, true)

		for i, amount := 0, len(sourcefields.tablerecords); i < amount; i++ {
			sourcefield := sourcefields.tablerecords[i]
			flags := parseFieldTags(sourcefield.structField.Tag)
			accessor, ok := targetIterator.Next()
			if flags.isFlagExists(Ignore) || !ok {
				continue
			}
			srcval, dstval := sourcefield.FieldValue(), accessor.FieldValue()
			if srcval == nil || dstval == nil {
				t.Logf("%d. field info missed", i)
				continue
			}
			t.Logf("%d. %s (%v) -> %s (%v) | %v", i, strings.Join(reverseStringSlice(sourcefield.names), "."), valfmt(srcval), accessor.StructFieldName(), valfmt(dstval), typfmt(accessor.StructField().Type))
			// ec.Attach(invokeStructFieldTransformer(c, params, srcval, dstval, padding))
		}

	})

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

func testgetallfieldsX1(t *testing.T) {

	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := rdecode(v1)

	var sourcefields fieldstable
	sourcefields = sourcefields.getallfields(t1, false)
	for i, amount := 0, len(sourcefields.tablerecords); i < amount; i++ {
		sourcefield := sourcefields.tablerecords[i]
		srcval := sourcefield.FieldValue()
		srctypstr := typfmtv(srcval)
		functorLog("%d. %s (%v) %v -> %s (%v)", i, strings.Join(reverseStringSlice(sourcefield.names), "."), valfmt(srcval), srctypstr, "", "")
	}
}

func teststructiteratorNextT1(t *testing.T) {

	x1 := x1data()
	v1 := reflect.ValueOf(&x1)
	t1, _ := rdecode(v1)

	it := newStructIterator(t1)
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		functorLog("%d. %q (%v) %v %q", i, field.Name, typfmt(field.Type), field.Index, field.PkgPath)
	}
}

func TestStructIterator_Next_Employee2(t *testing.T) {
	t.Run("testStructIterator_Next_Employee2", teststructiteratorNextEmployee2)
	t.Run("testStructIterator_Next_Employee2_exp", teststructiteratorNextEmployee2Exp)
}

func teststructiteratorNextEmployee2(t *testing.T) {

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
	t1, _ := rdecode(v1)
	it := newStructIterator(t1)
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		t.Logf("%d. %q (%v) %v %q", i, field.Name, typfmt(field.Type), field.Index, field.PkgPath)
		sb.WriteString(fmt.Sprintf("%d. %q (%v) %v %q\n", i, field.Name, typfmt(field.Type), field.Index, field.PkgPath))
	}

	if sb.String() != `0. "Base" (deepcopy.Base (struct)) [0] ""
1. "Avatar" (string (string)) [1] ""
2. "Image" ([]uint8 (slice)) [2] ""
3. "Attr" (*deepcopy.Attr (ptr)) [3] ""
4. "Valid" (bool (bool)) [4] ""
5. "Deleted" (bool (bool)) [5] ""
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}
}

func teststructiteratorNextEmployee2Exp(t *testing.T) {

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
	t1, _ := rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		if field == nil {
			t.Logf("%d. field info missed", i)
			continue
		}
		t.Logf("%d. %q (%v) %v %q", i, field.Name, typfmt(field.Type), field.Index, field.PkgPath)
		sb.WriteString(fmt.Sprintf("%d. %q (%v) %v %q\n", i, field.Name, typfmt(field.Type), field.Index, field.PkgPath))
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
	t.Run("testStructIterator_Next_User_new", teststructiteratorNextUserNew)

	t.Run("testStructIterator_Next_User", teststructiteratorNextUser)
	t.Run("testStructIterator_Next_User_zero", teststructiteratorNextUserZero)
	t.Run("testStructIterator_Next_User_more", teststructiteratorNextUserMore)
}

func teststructiteratorNextUser(t *testing.T) {

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
	t1, _ := rdecode(v1)
	it := newStructIterator(t1)
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v)%v | fld: %+v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index, field)
	}

}

func teststructiteratorNextUserNew(t *testing.T) {

	//timeZone, _ := time.LoadLocation("America/Phoenix")
	//timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	////tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	//tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	////tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	//tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, tgt := range []*User{
		new(User),
	} {
		sb.WriteString("\n")
		v1 := reflect.ValueOf(&tgt)
		t1, _ := rdecode(v1)
		it := newStructIterator(t1, withStructPtrAutoExpand(true), withStructFieldPtrAutoNew(true))
		for i := 0; ; i++ {
			accessor, ok := it.Next()
			if !ok {
				break
			}
			field := accessor.StructField()
			_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v)%v | fld: %+v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index, field)
		}
	}

}

func teststructiteratorNextUserZero(t *testing.T) {

	//timeZone, _ := time.LoadLocation("America/Phoenix")
	//timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	//tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	//tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	//tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	//tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, tgt := range []*User{
		new(User),
	} {
		sb.WriteString("\n\n\n")
		v1 := reflect.ValueOf(&tgt)
		t1, _ := rdecode(v1)
		it := newStructIterator(t1)
		for i := 0; ; i++ {
			accessor, ok := it.Next()
			if !ok {
				break
			}
			field := accessor.StructField()
			_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v)%v | fld: %+v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index, field)
		}
	}

}

func teststructiteratorNextUserMore(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	//tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	//tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

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
		v1 := reflect.ValueOf(&tgt)
		t1, _ := rdecode(v1)
		it := newStructIterator(t1)
		for i := 0; ; i++ {
			accessor, ok := it.Next()
			if !ok {
				break
			}
			field := accessor.StructField()
			_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v(%v)%v | fld: %+v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index, field)
		}
	}

}

func TestStructIterator_Next_A4(t *testing.T) {
	t.Run("testStructIterator_Next_A4_new", teststructiteratorNextA4New)

	t.Run("testStructIterator_Next_A4_zero", teststructiteratorNextA4Zero)
	t.Run("testStructIterator_Next_A4", teststructiteratorNextA4)
}

func teststructiteratorNextA4New(t *testing.T) {

	a4 := new(A4)

	var sb strings.Builder

	v1 := reflect.ValueOf(&a4)
	t1, _ := rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true), withStructFieldPtrAutoNew(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v (%v) %v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index)
	}

	t.Logf(sb.String())
	if sb.String() != `0. "Name2" (string (string)) | deepcopy.A2 (struct) (0) [0]
1. "Int2" (int (int)) | deepcopy.A2 (struct) (1) [1]
2. "Bool2" (bool (bool)) | deepcopy.A2 (struct) (2) [2]
3. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
4. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
5. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
6. "Name3" (string (string)) | deepcopy.A3 (struct) (1) [1]
7. "Int3" (int (int)) | deepcopy.A3 (struct) (2) [2]
8. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
9. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
10. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
11. "Bool3" (bool (bool)) | deepcopy.A3 (struct) (4) [4]
12. "Int4" (int (int)) | deepcopy.A4 (struct) (1) [1]
13. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
14. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
15. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. "Name2" (string (string)) | deepcopy.A2 (struct) (0) [0]
	// 1. "Int2" (int (int)) | deepcopy.A2 (struct) (1) [1]
	// 2. "Bool2" (bool (bool)) | deepcopy.A2 (struct) (2) [2]
	// 3. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 4. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 5. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
	// 6. "Name3" (string (string)) | deepcopy.A3 (struct) (1) [1]
	// 7. "Int3" (int (int)) | deepcopy.A3 (struct) (2) [2]
	// 8. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 9. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 10. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
	// 11. "Bool3" (bool (bool)) | deepcopy.A3 (struct) (4) [4]
	// 12. "Int4" (int (int)) | deepcopy.A4 (struct) (1) [1]
	// 13. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 14. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 15. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
}

func teststructiteratorNextA4Zero(t *testing.T) {

	a4 := new(A4)

	var sb strings.Builder

	v1 := reflect.ValueOf(&a4)
	t1, _ := rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v (%v) %v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index)
	}

	t.Logf(sb.String())
	if sb.String() != `0. "Name2" (string (string)) | deepcopy.A2 (struct) (0) [0]
1. "Int2" (int (int)) | deepcopy.A2 (struct) (1) [1]
2. "Bool2" (bool (bool)) | deepcopy.A2 (struct) (2) [2]
3. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
4. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
5. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
6. "Name3" (string (string)) | deepcopy.A3 (struct) (1) [1]
7. "Int3" (int (int)) | deepcopy.A3 (struct) (2) [2]
8. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
9. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
10. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
11. "Bool3" (bool (bool)) | deepcopy.A3 (struct) (4) [4]
12. "Int4" (int (int)) | deepcopy.A4 (struct) (1) [1]
13. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
14. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
15. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. "Name2" (string (string)) | deepcopy.A2 (struct) (0) [0]
	// 1. "Int2" (int (int)) | deepcopy.A2 (struct) (1) [1]
	// 2. "Bool2" (bool (bool)) | deepcopy.A2 (struct) (2) [2]
	// 3. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 4. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 5. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
	// 6. "Name3" (string (string)) | deepcopy.A3 (struct) (1) [1]
	// 7. "Int3" (int (int)) | deepcopy.A3 (struct) (2) [2]
	// 8. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 9. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 10. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
	// 11. "Bool3" (bool (bool)) | deepcopy.A3 (struct) (4) [4]
	// 12. "Int4" (int (int)) | deepcopy.A4 (struct) (1) [1]
	// 13. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 14. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 15. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
}

func teststructiteratorNextA4(t *testing.T) {

	a4 := prepareDataA4()

	var sb strings.Builder

	v1 := reflect.ValueOf(&a4)
	t1, _ := rdecode(v1)
	it := newStructIterator(t1, withStructPtrAutoExpand(true))
	for i := 0; ; i++ {
		accessor, ok := it.Next()
		if !ok {
			break
		}
		field := accessor.StructField()
		_, _ = fmt.Fprintf(&sb, "%d. %q (%v) | %v (%v) %v\n", i, field.Name, typfmt(field.Type), typfmt(accessor.Type()), accessor.index, field.Index)
	}

	t.Logf(sb.String())
	if sb.String() != `0. "Name2" (string (string)) | deepcopy.A2 (struct) (0) [0]
1. "Int2" (int (int)) | deepcopy.A2 (struct) (1) [1]
2. "Bool2" (bool (bool)) | deepcopy.A2 (struct) (2) [2]
3. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
4. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
5. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
6. "Name3" (string (string)) | deepcopy.A3 (struct) (1) [1]
7. "Int3" (int (int)) | deepcopy.A3 (struct) (2) [2]
8. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
9. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
10. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
11. "Bool3" (bool (bool)) | deepcopy.A3 (struct) (4) [4]
12. "Int4" (int (int)) | deepcopy.A4 (struct) (1) [1]
13. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
14. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
15. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
` {
		t.Fail()
	} else {
		t.Log("The loops and outputs verified")
	}

	// Output:
	// 0. "Name2" (string (string)) | deepcopy.A2 (struct) (0) [0]
	// 1. "Int2" (int (int)) | deepcopy.A2 (struct) (1) [1]
	// 2. "Bool2" (bool (bool)) | deepcopy.A2 (struct) (2) [2]
	// 3. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 4. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 5. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
	// 6. "Name3" (string (string)) | deepcopy.A3 (struct) (1) [1]
	// 7. "Int3" (int (int)) | deepcopy.A3 (struct) (2) [2]
	// 8. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 9. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 10. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
	// 11. "Bool3" (bool (bool)) | deepcopy.A3 (struct) (4) [4]
	// 12. "Int4" (int (int)) | deepcopy.A4 (struct) (1) [1]
	// 13. "Name1" (string (string)) | deepcopy.A1 (struct) (0) [0]
	// 14. "Int1" (int (int)) | deepcopy.A1 (struct) (1) [1]
	// 15. "Bool1" (bool (bool)) | deepcopy.A1 (struct) (2) [2]
}

func TestFieldsTable_getallfields(t *testing.T) {
	t.Run("test_getallfields X1 deep", testfieldstableGetFieldsDeeply)
	t.Run("test_getallfields", testfieldstableGetallfields)
	t.Run("test_getallfields_Employee2", testfieldstableGetallfieldsEmployee2)
	t.Run("test_getallfields_Employee2_2", testfieldstableGetallfieldsEmployee22)
	t.Run("test_getallfields_User", testfieldstableGetallfieldsUser)
}

func testfieldstableGetFieldsDeeply(t *testing.T) {

	x2 := new(X1)
	t2 := rdecodesimple(reflect.ValueOf(&x2))

	var sourcefields fieldstable
	sourcefields.getallfields(t2, true)

	var sb strings.Builder
	defer func() { t.Log(sb.String()) }()

	for i, f := range sourcefields.tablerecords {
		_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
			f.FieldName(), f.indexes,
			typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
		)
	}

}

func testfieldstableGetallfields(t *testing.T) {
	a4 := prepareDataA4()
	v4 := reflect.ValueOf(&a4)

	var sourcefields fieldstable
	sourcefields.getallfields(v4, true)

	var sb strings.Builder
	defer func() { t.Log(sb.String()) }()

	for i, f := range sourcefields.tablerecords {
		_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
			f.FieldName(), f.indexes,
			typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
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
	//0. [Name2 A2], [0 0] | &{Name2  string  0 [0] false}
	//1. [Int2 A2], [1 0] | &{Int2  int  16 [1] false}
	//2. [Bool2 A2], [2 0] | &{Bool2  bool  24 [2] false}
	//3. [Name1 A1], [0 3] | &{Name1  string  0 [0] false}
	//4. [Int1 A1], [1 3] | &{Int1  int  16 [1] false}
	//5. [Bool1 A1], [2 3] | &{Bool1  bool  24 [2] false}
	//6. [Name3 A3], [1 0] | &{Name3  string  8 [1] false}
	//7. [Int3 A3], [2 0] | &{Int3  int  24 [2] false}
	//8. [Name1 A1], [0 3] | &{Name1  string  0 [0] false}
	//9. [Int1 A1], [1 3] | &{Int1  int  16 [1] false}
	//10. [Bool1 A1], [2 3] | &{Bool1  bool  24 [2] false}
	//11. [Bool3 A3], [4 0] | &{Bool3  bool  64 [4] false}
	//12. [Int4], [1] | &{Int4  int  8 [1] false}
	//13. [Name1 A1], [0 2] | &{Name1  string  0 [0] false}
	//14. [Int1 A1], [1 2] | &{Int1  int  16 [1] false}
	//15. [Bool1 A1], [2 2] | &{Bool1  bool  24 [2] false}
}

func testfieldstableGetallfieldsEmployee2(t *testing.T) {

	//timeZone, _ := time.LoadLocation("America/Phoenix")
	//timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	//tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	//tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	//tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	//tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	var sb strings.Builder
	defer func() {
		t.Logf("\n%v\n", sb.String())
	}()

	for _, a4 := range []*Employee2{
		new(Employee2),
		//{
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
		//},
	} {
		sb.WriteString("\n")

		v4 := reflect.ValueOf(&a4)

		var sourcefields fieldstable
		sourcefields.getallfields(v4, true)

		var sb strings.Builder
		defer func() { t.Log(sb.String()) }()

		for i, f := range sourcefields.tablerecords {
			_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
				f.FieldName(), f.indexes,
				typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
			)
		}

		if sb.String() != `0. Base.Name, [0 0] | string (string), "", ""
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

func testfieldstableGetallfieldsEmployee22(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	//timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	//tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	//tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	//tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	var sb strings.Builder
	defer func() { t.Logf("\n%v\n", sb.String()) }()

	for _, a4 := range []*Employee2{
		//new(Employee2),
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

		v4 := reflect.ValueOf(&a4)

		var sourcefields fieldstable
		sourcefields.getallfields(v4, true)

		for i, f := range sourcefields.tablerecords {
			_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
				f.FieldName(), f.indexes,
				typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
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

func testfieldstableGetallfieldsUser(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	//tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	//tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
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
		//sb.WriteString("\n")

		v4 := reflect.ValueOf(&a4)

		var sourcefields fieldstable
		sourcefields.getallfields(v4, true)

		for i, f := range sourcefields.tablerecords {
			_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v, %q, %q\n", i,
				f.FieldName(), f.indexes,
				typfmt(f.structField.Type), f.structField.Tag, f.structField.PkgPath,
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
