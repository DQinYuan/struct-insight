package insight

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"github.com/tylerb/gls"
)

/*
|-k:v 表示结构体字段
--v   表示数组的元素
--k:v 表示map的元素

name(type)  表示复合的结构体,匿名复合结构体则type为空

空指针全部跳过


(Web)
|-Host:xxx
|-port:xxx
|-Timeout:xxx
|-Rate:xxx
|-Score
  --10
  --23
  --00
|-Ip:
  --xxx
  --yyy
|-MySQL:()
  |-Name:xxx
  |-port:9876
|-redis:(Redis)
  |-name:xxx
  |-port:6379
  |-host:(Host)
    |-ip:xxx
    |-alias
      --xxx
      --xxx
    |-policy(Policy)
      |-Allow:xxxx
      |-Deny:xxxx


insight.Start('key')
insight.Insight('comment', object)

 */

const (
	NON = iota
	OBJ
	ARR_MAP
)

var prefixes = []string{"", "|-", "--"}

func addSpace(buf *bytes.Buffer, dep int) {
	spaces := make([]byte, dep*2)
	for i := range spaces {
		spaces[i] = 32
	}

	buf.Write(spaces)
}

func push(values *[]interface{}, v interface{})  {
	*values = append(*values, v)
}

func prettify(v reflect.Value, buf *bytes.Buffer, values *[]interface{}, dep int, prefixType int,
	key string) {
	// 防止 zero value  和 不公开属性
	if !v.IsValid() || !v.CanInterface() {
		return
	}
	k := v.Kind()
	// 指针类型不会增加深度
	if k == reflect.Ptr {
		prettify(v.Elem(), buf, values, dep, prefixType, key)
		return
	}

	addSpace(buf, dep)
	buf.WriteString(prefixes[prefixType])
	if key != "" {
		buf.WriteString("%s:")
		push(values, key)
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		buf.WriteString("%d\n")
	    push(values, v.Interface())
	case reflect.Bool:
		buf.WriteString("%t\n")
		push(values, v.Interface())
	case reflect.Int64:
		if v.Type().String() == "time.Duration" {
			buf.WriteString("%v\n")
		} else {
			buf.WriteString("%d\n")
		}
		push(values, v.Interface())
	case reflect.String:
		buf.WriteString("%s\n")
		push(values, v.Interface())
	case reflect.Float32, reflect.Float64:
		buf.WriteString("%f\n")
		push(values, v.Interface())
	case reflect.Struct:
		buf.WriteString("(%s)\n")
		push(values, v.Type().String())
		for i := 0; i < v.NumField(); i++ {
			prettify(v.Field(i), buf, values, dep + 1, OBJ, v.Type().Field(i).Name)
		}
	case reflect.Slice, reflect.Array:
		buf.WriteString("\n")
		for i := 0; i < v.Len(); i++ {
			prettify(v.Index(i), buf, values, dep + 1, ARR_MAP, "")
		}
	case reflect.Map:
		buf.WriteString("\n")
		for _, key := range v.MapKeys() {
			prettify(v.MapIndex(key), buf, values, dep + 1, ARR_MAP,
				fmt.Sprintf("%v", key.Interface()))
		}
	default:
		buf.WriteString("<not known type>\n")
	}
}

const FILEKEY = "file"

func Start(key string) {
	f, err := os.OpenFile(key, os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		log.Printf("open key file %s error %v\n", key, err)
		return
	}

	gls.Set(FILEKEY, f)
}

func Insight(comment string, v interface{}) {
	buf := bytes.NewBufferString(comment + " :\n")
	values := make([]interface{}, 0)

	prettify(reflect.ValueOf(v), buf, &values, 0, NON, "")

	result := fmt.Sprintf(buf.String(), values...)

	fmt.Println(result)

	if gls.Get(FILEKEY) != nil {
		gls.Get(FILEKEY).(*os.File).WriteString(result)
	}
}

