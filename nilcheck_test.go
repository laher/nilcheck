package nilcheck

import "testing"

type MyInterface interface{}

type MyStruct1 struct {
	Name        string
	Number      int
	OtherNumber float64
}

type MyStruct2 struct {
	Struct             MyStruct1
	Pointer            *MyStruct1
	MapStringString    map[string]string
	MapStringInterface map[string]interface{}
	Slice              []string
	MyInterface        MyInterface
}
type MyStruct3 struct {
	Channel chan int
}

var z = &MyStruct3{}

var nilStruct *MyStruct1
var nilPrim *int
var nilMap map[string]string = nil
var nilSlice []string = nil
var arrayWithNils = [1]*int{}
var arrayWithZeroLength = [0]*int{}

var hasNoNilsTests = []interface{}{
	MyStruct1{},
	MyStruct3{Channel: make(chan int)},
	6,
	"hello",
	map[string]string{},
	arrayWithZeroLength,
}

var hasNilTests = []interface{}{
	nil,
	nilStruct,
	nilPrim,
	nilMap,
	nilSlice,
	arrayWithNils,
	MyStruct3{},
	map[string]interface{}{
		"y": z,
	},
	map[string]*int{
		"y": nil,
	},
	map[*int]string{
		nil: "x",
	},
	struct {
		MyStruct2 MyStruct2
	}{
		MyStruct2: MyStruct2{
			Pointer:         &MyStruct1{},
			MapStringString: map[string]string{},
			MapStringInterface: map[string]interface{}{
				//				"x": nil,
				"y": z,
			},
			MyInterface: z,
		},
	},
}

func TestNilcheck1(t *testing.T) {
	nc := NewNilChecker()

	for _, hasNilTest := range hasNilTests {
		err := nc.Check(hasNilTest)
		if err == nil {
			t.Errorf("Test should find a nil. Input: [%#v]", hasNilTest)
		}
	}

	for _, testObj := range hasNoNilsTests {
		err := nc.Check(testObj)
		if err != nil {
			t.Errorf("Test should NOT find a nil. Error: [%s]. Input: [%#v]", err, testObj)
		}
	}
}
