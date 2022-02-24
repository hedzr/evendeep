package deepcopy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestStructIterator_Next_X1(t *testing.T) {

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

	v1 := reflect.ValueOf(&x1)
	t1, _ := rdecode(v1)

	t.Run("getallfields at once", func(t *testing.T) {

		var sourcefields fieldstable
		sourcefields = sourcefields.getallfields(t1, false)
		for i, amount := 0, len(sourcefields.records); i < amount; i++ {
			sourcefield := sourcefields.records[i]
			srcval := sourcefield.Value()
			srctypstr := typfmtv(srcval)
			functorLog("%d. %s (%v) %v -> %s (%v)", i, strings.Join(reverseStringSlice(sourcefield.names), "."), valfmt(srcval), srctypstr, "", "")
		}

	})

	t.Run("by struct iterator", func(t *testing.T) {

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

	})

	t.Run("get further: loop src & dst at same time", func(t *testing.T) {

		targetIterator := newStructIterator(t1)

		var sourcefields fieldstable
		sourcefields = sourcefields.getallfields(t1, false)
		for i, amount := 0, len(sourcefields.records); i < amount; i++ {
			sourcefield := sourcefields.records[i]
			flags := parseFieldTags(sourcefield.structField.Tag)
			if flags.isFlagExists(Ignore) {
				continue
			}
			accessor, ok := targetIterator.Next()
			if !ok {
				continue
			}
			srcval, dstval := sourcefield.Value(), accessor.FieldValue()
			if srcval == nil || dstval == nil {
				t.Logf("%d. field info missed", i)
				continue
			}
			functorLog("%d. %s (%v) -> %s (%v) | %v", i, strings.Join(reverseStringSlice(sourcefield.names), "."), valfmt(srcval), accessor.StructFieldName(), valfmt(dstval), typfmt(accessor.StructField().Type))
			// ec.Attach(invokeStructFieldTransformer(c, params, srcval, dstval, padding))
		}

	})

}

func TestStructIterator_Next_Employee2(t *testing.T) {

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

	if sb.String() != `0. "Name" (string (string)) [0] ""
1. "Birthday" (*time.Time (ptr)) [1] ""
2. "Age" (int (int)) [2] ""
3. "EmployeID" (int64 (int64)) [3] ""
4. "Avatar" (string (string)) [1] ""
5. "Image" ([]uint8 (slice)) [2] ""
6. "Attr" (*deepcopy.Attr (ptr)) [3] ""
7. "Valid" (bool (bool)) [4] ""
8. "Deleted" (bool (bool)) [5] ""
` {
		t.Fail()
	}
}

func TestStructIterator_Next_Employee2_exp(t *testing.T) {

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
	}
}

func TestStructIterator_Next_User(t *testing.T) {

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

func TestStructIterator_Next_A4(t *testing.T) {

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
	a4 := prepareDataA4()
	v4 := reflect.ValueOf(&a4)

	var sourcefields fieldstable
	sourcefields.getallfields(v4, true)

	var sb strings.Builder
	for i, f := range sourcefields.records {
		_, _ = fmt.Fprintf(&sb, "%v. %v, %v | %v\n", i, f.names, f.indexes, f.structField)
	}

	t.Log(sb.String())

	if sb.String() != `0. [Name2 A2], [0 0] | &{Name2  string  0 [0] false}
1. [Int2 A2], [1 0] | &{Int2  int  16 [1] false}
2. [Bool2 A2], [2 0] | &{Bool2  bool  24 [2] false}
3. [Name1 A1], [0 3] | &{Name1  string  0 [0] false}
4. [Int1 A1], [1 3] | &{Int1  int  16 [1] false}
5. [Bool1 A1], [2 3] | &{Bool1  bool  24 [2] false}
6. [Name3 A3], [1 0] | &{Name3  string  8 [1] false}
7. [Int3 A3], [2 0] | &{Int3  int  24 [2] false}
8. [Name1 A1], [0 3] | &{Name1  string  0 [0] false}
9. [Int1 A1], [1 3] | &{Int1  int  16 [1] false}
10. [Bool1 A1], [2 3] | &{Bool1  bool  24 [2] false}
11. [Bool3 A3], [4 0] | &{Bool3  bool  64 [4] false}
12. [Int4], [1] | &{Int4  int  8 [1] false}
13. [Name1 A1], [0 2] | &{Name1  string  0 [0] false}
14. [Int1 A1], [1 2] | &{Int1  int  16 [1] false}
15. [Bool1 A1], [2 2] | &{Bool1  bool  24 [2] false}
` {
		t.Fail()
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
