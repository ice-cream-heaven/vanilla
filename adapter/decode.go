package adapter

import (
	"fmt"
	"github.com/ice-cream-heaven/log"
	"reflect"
	"strings"
)

func decodeSlice(dst []any, src any) error {
	t := reflect.TypeOf(src)
	if t.Kind() != reflect.Slice {
		panic("src is not map")
	}

	v := reflect.ValueOf(src)

	for i := 0; i < v.Len(); i++ {
		lv := v.Index(i)

		switch lv.Kind() {
		case reflect.Bool:
			dst = append(dst, lv.Bool())
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			dst = append(dst, lv.Int())
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			dst = append(dst, lv.Uint())
		case reflect.Float32, reflect.Float64:
			dst = append(dst, lv.Float())
		case reflect.Complex64, reflect.Complex128:
			dst = append(dst, lv.Complex())
		case reflect.Interface:
			dst = append(dst, lv.Interface())
		case reflect.Map:
			m := map[string]any{}
			dst = append(dst, m)
			err := decodeMap(m, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.Slice:
			var l []any
			dst = append(dst, l)
			err := decodeSlice(l, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.String:
			dst = append(dst, lv.String())
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			dst = append(dst, m)
			err := decode(m, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		default:
			log.Debugf("unknown kind %s", lv.Kind())
		}
	}

	return nil
}

func decodeMap(dst map[string]any, src any) error {
	t := reflect.TypeOf(src)
	if t.Kind() != reflect.Map {
		panic("src is not map")
	}

	v := reflect.ValueOf(src)

	for _, mk := range v.MapKeys() {
		mv := v.MapIndex(mk)
		mk := fmt.Sprintf("%v", mk.Interface())

		switch mv.Kind() {
		case reflect.Bool:
			dst[mk] = mv.Bool()
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			dst[mk] = mv.Int()
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			dst[mk] = mv.Uint()
		case reflect.Float32, reflect.Float64:
			dst[mk] = mv.Float()
		case reflect.Complex64, reflect.Complex128:
			dst[mk] = mv.Complex()
		case reflect.Interface:
			dst[mk] = mv.Interface()
		case reflect.Map:
			m := map[string]any{}
			dst[mk] = m
			err := decodeMap(m, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.Slice:
			var l []any
			dst[mk] = l
			err := decodeSlice(l, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.String:
			dst[mk] = mv.String()
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			dst[mk] = m
			err := decode(m, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		default:
			log.Debugf("unknown kind %s", mv.Kind())
		}
	}

	return nil
}

func decode(dst map[string]any, src any) error {
	if src == nil {
		return nil
	}

	for reflect.TypeOf(src).Kind() == reflect.Ptr {
		src = reflect.ValueOf(src).Elem().Interface()
	}

	t := reflect.TypeOf(src)
	v := reflect.ValueOf(src)

	for idx := 0; idx < t.NumField(); idx++ {
		ft := t.Field(idx)
		fv := v.Field(idx)

		tag := strings.TrimSuffix(ft.Tag.Get("proxy"), ",omitempty")

		switch fv.Kind() {
		case reflect.Bool:
			dst[tag] = fv.Bool()
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			dst[tag] = fv.Int()
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			dst[tag] = fv.Uint()
		case reflect.Float32, reflect.Float64:
			dst[tag] = fv.Float()
		case reflect.Complex64, reflect.Complex128:
			dst[tag] = fv.Complex()
		case reflect.Interface:
			dst[tag] = fv.Interface()
		case reflect.Map:
			m := map[string]any{}
			dst[tag] = m
			err := decodeMap(m, fv.Interface())
			if err != nil {
				return err
			}
		case reflect.Slice:
			var l []any
			dst[tag] = l
			err := decodeSlice(l, fv.Interface())
			if err != nil {
				return err
			}
		case reflect.String:
			dst[tag] = fv.String()
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			dst[tag] = m
			err := decode(m, fv.Interface())
			if err != nil {
				return err
			}
		default:
			log.Debugf("unknown kind %s", fv.Kind())
		}
	}

	return nil
}
