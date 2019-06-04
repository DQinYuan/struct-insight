package insight

import (
	"bytes"
	"fmt"
	"github.com/tylerb/gls"
	"log"
	"os"
	"reflect"
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

map类型的显示会受私有影响

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
	// 防止 zero value
	if !v.IsValid() /*|| !v.CanInterface()*/ {
		return
	}

	// 排除字段
	if gls.Get(MUSTEXCLUDE) != nil {
		excludeSet := gls.Get(MUSTEXCLUDE).(map[string]bool)
		if excludeSet[key] || excludeSet[v.Type().String()] {
			return
		}
	}

	// 私有字段
	if gls.Get(MUSTVISIT) != nil {
		visitSet := gls.Get(MUSTVISIT).(map[string]bool)
		if !v.CanInterface() && !visitSet[v.Type().String()] && !visitSet[key] {
			return
		}
	}

	if dep > 20 {
		return
	}
	k := v.Kind()
	// 指针类型或者接口类型不会增加深度
	if k == reflect.Ptr || k == reflect.Interface {
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		buf.WriteString("%d\n")
	    push(values, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf.WriteString("%d\n")
		push(values, v.Uint())
	case reflect.Bool:
		buf.WriteString("%t\n")
		push(values, v.Bool())
	case reflect.Int64:
		if v.CanInterface() && v.Type().String() == "time.Duration" {
			buf.WriteString("%v\n")
			timeStr := v.MethodByName("String").Call([]reflect.Value{})[0].Interface()
			push(values, timeStr)
		} else {
			buf.WriteString("%d\n")
			push(values, v.Int())
		}
	case reflect.String:
		buf.WriteString("%s\n")
		push(values, v.String())
	case reflect.Float32, reflect.Float64:
		buf.WriteString("%f\n")
		push(values, v.Float())
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
			if key.CanInterface() {
				prettify(v.MapIndex(key), buf, values, dep + 1, ARR_MAP,
					fmt.Sprintf("%v", key.Interface()))
			}
		}
	default:
		buf.WriteString("<not known type>\n")
	}
}

const FILEKEY = "file"
const MUSTVISIT = "mustvisit"
const MUSTEXCLUDE  = "mustexclude"


func Start(key string, mustvisit map[string]bool, mustexclude map[string]bool) {
	f, err := os.OpenFile(key, os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		log.Printf("open key file %s error %v\n", key, err)
		return
	}

	gls.Set(FILEKEY, f)

	if mustvisit != nil {
		gls.Set(MUSTVISIT, mustvisit)
	}

	if mustexclude != nil {
		gls.Set(MUSTEXCLUDE, mustexclude)
	}
}

func Insight(comment string, v interface{}) {
	if gls.Get(FILEKEY) == nil {
		return
	}

	buf := bytes.NewBufferString(comment + " :\n")
	values := make([]interface{}, 0)

	prettify(reflect.ValueOf(v), buf, &values, 0, NON, "")

	result := fmt.Sprintf(buf.String(), values...)

	fmt.Println(result)

	if gls.Get(FILEKEY) != nil {
		gls.Get(FILEKEY).(*os.File).WriteString(result)
	}
}

