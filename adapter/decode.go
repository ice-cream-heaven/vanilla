package adapter

import (
	"encoding/base64"
	"fmt"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/ice-cream-heaven/log"
	"reflect"
	"strings"
)

var whiteTag = map[string]bool{
	"alterId": true,
}

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
			if !lv.Bool() {
				continue
			}
			dst = append(dst, lv.Bool())
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			if lv.Int() == 0 {
				continue
			}
			dst = append(dst, lv.Int())
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			if lv.Uint() == 0 {
				continue
			}
			dst = append(dst, lv.Uint())
		case reflect.Float32, reflect.Float64:
			if lv.Float() == 0 {
				continue
			}
			dst = append(dst, lv.Float())
		case reflect.Complex64, reflect.Complex128:
			dst = append(dst, lv.Complex())
		case reflect.Interface:
			if lv.Interface() == nil {
				continue
			}
			dst = append(dst, lv.Interface())
		case reflect.Map:
			m := map[string]any{}
			err := decodeMap(m, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
			if len(m) == 0 {
				continue
			}
			dst = append(dst, m)
		case reflect.Slice:
			var l []any
			err := decodeSlice(l, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
			if len(l) == 0 {
				continue
			}
			dst = append(dst, l)
		case reflect.String:
			if lv.String() == "" {
				continue
			}
			dst = append(dst, lv.String())
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			err := decode(m, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
			if len(m) == 0 {
				continue
			}
			dst = append(dst, m)
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
			if !whiteTag[mk] && !mv.Bool() {
				continue
			}
			dst[mk] = mv.Bool()
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			if !whiteTag[mk] && mv.Int() == 0 {
				continue
			}
			dst[mk] = mv.Int()
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			if !whiteTag[mk] && mv.Uint() == 0 {
				continue
			}
			dst[mk] = mv.Uint()
		case reflect.Float32, reflect.Float64:
			if !whiteTag[mk] && mv.Float() == 0 {
				continue
			}
			dst[mk] = mv.Float()
		case reflect.Complex64, reflect.Complex128:
			dst[mk] = mv.Complex()
		case reflect.Interface:
			dst[mk] = mv.Interface()
		case reflect.Map:
			m := map[string]any{}
			err := decodeMap(m, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
			if !whiteTag[mk] && len(m) == 0 {
				continue
			}
			dst[mk] = m
		case reflect.Slice:
			var l []any
			err := decodeSlice(l, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
			if !whiteTag[mk] && len(l) == 0 {
				continue
			}
			dst[mk] = l
		case reflect.String:
			if !whiteTag[mk] && mv.String() == "" {
				continue
			}
			dst[mk] = mv.String()
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			err := decode(m, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
			if !whiteTag[mk] && len(m) == 0 {
				continue
			}
			dst[mk] = m
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
		if reflect.ValueOf(src).IsNil() {
			return nil
		}
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
			if !whiteTag[tag] && !fv.Bool() {
				continue
			}
			dst[tag] = fv.Bool()
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			if !whiteTag[tag] && fv.Int() == 0 {
				continue
			}
			dst[tag] = fv.Int()
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			if !whiteTag[tag] && fv.Uint() == 0 {
				continue
			}
			dst[tag] = fv.Uint()
		case reflect.Float32, reflect.Float64:
			if !whiteTag[tag] && fv.Float() == 0 {
				continue
			}
			dst[tag] = fv.Float()
		case reflect.Complex64, reflect.Complex128:
			dst[tag] = fv.Complex()
		case reflect.Interface:
			if !whiteTag[tag] && fv.IsNil() {
				continue
			}
			dst[tag] = fv.Interface()
		case reflect.Map:
			m := map[string]any{}
			err := decodeMap(m, fv.Interface())
			if err != nil {
				return err
			}
			if !whiteTag[tag] && len(m) == 0 {
				continue
			}
			dst[tag] = m
		case reflect.Slice:
			var l []any
			err := decodeSlice(l, fv.Interface())
			if err != nil {
				return err
			}
			if !whiteTag[tag] && len(l) == 0 {
				continue
			}
			dst[tag] = l
		case reflect.String:
			if fv.String() == "" {
				continue
			}
			dst[tag] = fv.String()
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			err := decode(m, fv.Interface())
			if err != nil {
				return err
			}
			if !whiteTag[tag] && len(m) == 0 {
				continue
			}
			dst[tag] = m
		default:
			log.Debugf("unknown kind %s", fv.Kind())
		}
	}

	return nil
}

func encode(src any) (map[string]any, error) {
	dst := map[string]any{}
	err := decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func ToOption(src any) any {
	if val, ok := src.(map[string]any); ok {
		if typ, ok := val["type"]; ok {
			var opt any
			switch typ {
			case "ss", "shadowsocks":
				opt = &outbound.ShadowSocksOption{}

			case "ssr", "shadowsocksr":
				opt = &outbound.ShadowSocksROption{}

			case "snell":
				opt = &outbound.SnellOption{}

			case "socks", "socks5", "socks4":
				opt = &outbound.Socks5Option{}

			case "http", "https":
				opt = &outbound.HttpOption{}

			case "vmess":
				opt = &outbound.VmessOption{}

			case "vless":
				opt = &outbound.VlessOption{}

			case "trojan":
				opt = &outbound.TrojanOption{}

			case "hysteria":
				opt = &outbound.HysteriaOption{}

			case "wireguard":
				opt = &outbound.WireGuardOption{}

			case "tuic":
				opt = &outbound.TuicOption{}

			default:
				return typ
			}

			err := decoder.Decode(val, opt)
			if err != nil {
				return err
			}

			return opt
		}

		return val
	} else {
		return src
	}
}

func Base64Decode(raw string) string {
	value, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		return string(value)
	}

	value, err = base64.URLEncoding.DecodeString(raw)
	if err == nil {
		return string(value)
	}

	value, err = base64.RawStdEncoding.DecodeString(raw)
	if err == nil {
		return string(value)
	}

	value, err = base64.RawURLEncoding.DecodeString(raw)
	if err == nil {
		return string(value)
	}

	return raw
}
