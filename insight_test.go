package insight

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type Expr interface {
	eval(v string) bool
}

type MyExpr struct {
	Haha string
}

func (*MyExpr) eval(v string) bool {
	return false
}

type Policy struct {
	Allow string
	Deny  string
}

type Host struct {
	Ip     string
	Alias  []string
	Policy Policy
}
type Redis struct {
	Name string // lower-case
	Port uint   // lower-case
	Host Host
}

type Web struct {
	Host    string
	port    int32 // lower-case
	Timeout time.Duration
	Rate    float32
	Score   []float32
	IP      []string
	Mymap   map[string]bool
	Epr     Expr
	MySQL   *struct {
		Name string
		port int64 // lower-case
	}
	Redis *Redis
}

func TestInsight(t *testing.T) {
	Start("test")
	w := &Web{
		Host:    "web host",
		port:    1234,
		Timeout: 5 * time.Second,
		Rate:    0.32,
		Score:   []float32{},
		IP:      []string{"192.168.1.1", "127.0.0.1", "localhost"},
		Mymap: map[string]bool{
			"aa":  true,
			"bbb": false,
		},
		Epr: &MyExpr{Haha:"iiiii"},
		MySQL: &struct {
			Name string
			port int64
		}{Name: "mysqldb", port: 3306},
		Redis: &Redis{"rdb", 6379, Host{"adf", []string{"alias1", "alias2"}, Policy{"allow policy", "deny policy"}}},
	}
	Insight("Web info", w)
}

func TestAscii(t *testing.T) {
	buf := bytes.NewBufferString("")
	buf.WriteString("aaaa")
	buf.Write([]byte{32, 32, 32})
	buf.WriteString("bbbbb")
	fmt.Println(buf.String())
}

func TestInterface(t *testing.T) {
	pol := &Policy{}
	of := reflect.ValueOf(pol)
	fmt.Println(of.CanInterface())
}
