package geerpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct{ Num1, Num2 int }

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Aum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)

	rcvr := reflect.ValueOf(&foo)
	name := reflect.Indirect(rcvr).Type().Name()
	typ := reflect.TypeOf(name)

	fmt.Printf("s.rcvr:%v \n", typ.NumMethod())

	_assert(len(s.method) == 0, "wrong service method, expect 1, but  got %d", len(s.method))
	mType := s.method["aum"]
	_assert(mType != nil, "wrong method, sum should not nil")
}

func TestMethodTypeCall(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo")
}
