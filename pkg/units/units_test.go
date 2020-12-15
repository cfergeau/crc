package units

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	units "github.com/docker/go-units"
)

func TestFromSize(t *testing.T) {
	assertSuccessEquals(t, 32, FromSize, "32")
	assertSuccessEquals(t, 32, FromSize, "32b")
	assertSuccessEquals(t, 32, FromSize, "32B")
	assertSuccessEquals(t, 32*units.KB, FromSize, "32k")
	assertSuccessEquals(t, 32*units.KB, FromSize, "32K")
	assertSuccessEquals(t, 32*units.KB, FromSize, "32kb")
	assertSuccessEquals(t, 32*units.KB, FromSize, "32Kb")
	assertSuccessEquals(t, 32*units.KiB, FromSize, "32Kib")
	assertSuccessEquals(t, 32*units.KiB, FromSize, "32KIB")
	assertSuccessEquals(t, 32*units.MB, FromSize, "32Mb")
	assertSuccessEquals(t, 32*units.GB, FromSize, "32Gb")
	assertSuccessEquals(t, 32*units.TB, FromSize, "32Tb")
	assertSuccessEquals(t, 32*units.PB, FromSize, "32Pb")
	assertSuccessEquals(t, 32*units.PB, FromSize, "32PB")
	assertSuccessEquals(t, 32*units.PB, FromSize, "32P")

	assertSuccessEquals(t, 32, FromSize, "32.3")
	tmp := 32.3 * units.MiB
	assertSuccessEquals(t, int64(tmp), FromSize, "32.3 MiB")

	assertError(t, FromSize, "")
	assertError(t, FromSize, "hello")
	assertError(t, FromSize, "-32")
	assertError(t, FromSize, " 32 ")
	assertError(t, FromSize, "32m b")
	assertError(t, FromSize, "32bm")
}

// func that maps to the parse function signatures as testing abstraction
type parseFn func(string) (int64, error)

// Define 'String()' for pretty-print
func (fn parseFn) String() string {
	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return fnName[strings.LastIndex(fnName, ".")+1:]
}

func assertSuccessEquals(t *testing.T, expected int64, fn parseFn, arg string) {
	res, err := fn(arg)
	if err != nil || res != expected {
		t.Errorf("%s(\"%s\") -> expected '%d' but got '%d' with error '%v'", fn, arg, expected, res, err)
	}
}

func assertError(t *testing.T, fn parseFn, arg string) {
	res, err := fn(arg)
	if err == nil && res != -1 {
		t.Errorf("%s(\"%s\") -> expected error but got '%d'", fn, arg, res)
	}
}
