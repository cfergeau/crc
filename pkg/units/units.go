package units

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	units "github.com/docker/go-units"
	"github.com/spf13/cast"
)

type Unit uint64

const (
	Bytes Unit = 1

	KB Unit = units.KB
	MB Unit = units.MB
	GB Unit = units.GB
	TB Unit = units.TB
	PB Unit = units.PB

	KiB Unit = units.KiB
	MiB Unit = units.MiB
	GiB Unit = units.GiB
	TiB Unit = units.TiB
	PiB Unit = units.PiB
)

type unitMap map[string]int64

var (
	decimalMap = unitMap{"k": units.KB, "m": units.MB, "g": units.GB, "t": units.TB, "p": units.PB}
	binaryMap  = unitMap{"k": units.KiB, "m": units.MiB, "g": units.GiB, "t": units.TiB, "p": units.PiB}
	sizeRegex  = regexp.MustCompile(`^(\d+(\.\d+)?) ?([kKmMgGtTpP])?([iI])?[bB]?$`)
)

// FromSize returns an integer from a specification of a
// size using either SI standard (eg. "44kB", "17MB") or
// binary standard (eg. "37kiB", "97MiB")
func FromSize(size string) (int64, error) {
	return parseSize(size, Auto)
}

type parsingMode int

const (
	Auto parsingMode = iota
	ForceBinary
	ForceDecimal
)

// Parses the size string into the amount it represents.
func parseSize(sizeStr string, mode parsingMode) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 5 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}
	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[3])

	var uMap unitMap
	switch mode {
	case ForceBinary:
		uMap = binaryMap
	case ForceDecimal:
		uMap = decimalMap
	case Auto:
		fallthrough
	default:
		if matches[4] != "" {
			uMap = binaryMap
		} else {
			uMap = decimalMap
		}
	}

	if mul, ok := uMap[unitPrefix]; ok {
		size *= float64(mul)
	}

	return int64(size), nil
}

type Size uint64

func (s Size) String() string {
	return fmt.Sprintf("%d", int64(s))
}

func (s Size) HumanSizeStr() string {
	return units.HumanSize(float64(s))
}

func (s Size) BytesSizeStr() string {
	return units.BytesSize(float64(s))
}

func (s Size) ConvertTo(unit Unit) uint64 {
	return uint64(s) / uint64(unit)
}

func (s Size) ToBytes() uint64 {
	return s.ConvertTo(Bytes)
}

func New(size uint64, unit Unit) Size {
	return Size(size * uint64(unit))
}

func ToSizeE(value interface{}, defaultUnit Unit) (Size, error) {
	rawValue := indirect(value)
	switch v := rawValue.(type) {
	case Size:
		return v, nil
	case string:
		bytes, err := FromSize(v)
		if err != nil {
			return Size(0), err
		}
		return New(uint64(bytes), Bytes), nil

	default:
		u, err := cast.ToUint64E(v)
		if err != nil {
			return Size(0), err
		}
		return New(u, defaultUnit), nil
	}
}

func ToSize(value interface{}, defaultUnit Unit) Size {
	size, _ := ToSizeE(value, defaultUnit)
	return size
}

// func indirect() is:
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
//
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
