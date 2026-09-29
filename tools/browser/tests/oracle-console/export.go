package main

// A private-media bridge to compare the new core against existing Go checkpoints.
// It follows the C++ field schema; it is not part of the deployed emulator.
import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type exporter struct {
	b      []byte
	fields map[string][]string
	ids    map[uintptr]uint32
}

func (e *exporter) u(v uint64, n int) {
	for i := 0; i < n; i++ {
		e.b = append(e.b, byte(v))
		v >>= 8
	}
}
func compare(a, b reflect.Value) int {
	switch a.Kind() {
	case reflect.String:
		return strings.Compare(a.String(), b.String())
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if n := compare(a.Field(i), b.Field(i)); n != 0 {
				return n
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() < b.Int() {
			return -1
		}
		if a.Int() > b.Int() {
			return 1
		}
	default:
		if a.Uint() < b.Uint() {
			return -1
		}
		if a.Uint() > b.Uint() {
			return 1
		}
	}
	return 0
}
func (e *exporter) value(v reflect.Value) {
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			e.u(0, 4)
			return
		}
		e.value(v.Elem())
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			e.u(0, 4)
			return
		}
		ptr := v.Pointer()
		id, known := e.ids[ptr]
		if !known {
			id = uint32(len(e.ids) + 1)
			e.ids[ptr] = id
		}
		e.u(uint64(id), 4)
		if known {
			e.u(0, 1)
			return
		}
		e.u(1, 1)
		e.value(v.Elem())
	case reflect.Bool:
		if v.Bool() {
			e.u(1, 1)
		} else {
			e.u(0, 1)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.u(v.Uint(), int(v.Type().Size()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.u(uint64(v.Int()), int(v.Type().Size()))
	case reflect.Float32:
		e.u(uint64(math.Float32bits(float32(v.Float()))), 4)
	case reflect.Float64:
		e.u(math.Float64bits(v.Float()), 8)
	case reflect.String:
		s := v.String()
		e.u(uint64(len(s)), 4)
		e.b = append(e.b, s...)
	case reflect.Slice:
		if v.IsNil() {
			e.u(0, 1)
		} else {
			e.u(1, 1)
		}
		e.u(uint64(v.Len()), 4)
		for i := 0; i < v.Len(); i++ {
			e.value(v.Index(i))
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			e.value(v.Index(i))
		}
	case reflect.Map:
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return compare(keys[i], keys[j]) < 0 })
		e.u(uint64(len(keys)), 4)
		for _, k := range keys {
			e.value(k)
			e.value(v.MapIndex(k))
		}
	case reflect.Struct:
		name := filepath.Base(v.Type().PkgPath()) + "_" + v.Type().Name()
		fields, ok := e.fields[name]
		if !ok {
			panic("no schema for " + name)
		}
		for _, f := range fields {
			if f == "rrVblAcc" || f == "rrIopAcc" {
				e.u(0, 8)
				continue
			}
			fv := v.FieldByName(strings.TrimSuffix(f, "_"))
			if !fv.IsValid() {
				panic(name + " missing " + f)
			}
			e.value(fv)
		}
	default:
		panic(fmt.Sprintf("cannot export %v", v.Type()))
	}
}
func exportState(platform, path string, m any) {
	raw, err := os.ReadFile("tools/platform/" + platform + "/browser/core/state-fields.h")
	if err != nil {
		panic(err)
	}
	e := exporter{fields: map[string][]string{}, ids: map[uintptr]uint32{}}
	re := regexp.MustCompile(`Archive&a,(\w+)&v\)\{a\(([^)]*)\);`)
	for _, r := range re.FindAllStringSubmatch(string(raw), -1) {
		if r[2] == "" {
			e.fields[r[1]] = nil
			continue
		}
		for _, f := range strings.Split(r[2], ",") {
			e.fields[r[1]] = append(e.fields[r[1]], strings.TrimPrefix(f, "v."))
		}
	}
	// Archive header matches archive.h.
	e.u(0x53525231, 4)
	id := uint32(11)
	if platform == "ps2" {
		id = 12
	}
	var h [8]byte
	version := uint32(2)
	if platform == "gba" {
		id = 13
		version = 1
	}
	if platform == "xbox" {
		id = 16
		version = 1
	}
	if platform == "dc" {
		id = 14
		version = 1
	}
	binary.LittleEndian.PutUint32(h[:4], id)
	binary.LittleEndian.PutUint32(h[4:], version)
	e.b = append(e.b, h[:]...)
	e.value(reflect.ValueOf(m))
	v := reflect.ValueOf(m).Elem()
	names := []string{"RAM", "ARAM"}
	if platform == "ps2" {
		names = []string{"ram"}
	}
	if platform == "xbox" {
		names = []string{"RAM"}
	}
	if platform == "gba" || platform == "dc" {
		names = nil
	}
	for _, n := range names {
		s := v.FieldByName(n)
		for i := 0; i < s.Len(); i++ {
			e.b = append(e.b, byte(s.Index(i).Uint()))
		}
	}
	if platform == "xbox" {
		e.u(0, 8)
	}
	if err = os.WriteFile(path, e.b, 0600); err != nil {
		panic(err)
	}
}
