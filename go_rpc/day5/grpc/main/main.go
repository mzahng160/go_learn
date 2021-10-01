package main

import (
	"context"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int
type Args struct{ Num1, Num2 int }

func (f Foo) Aum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addrCh chan string) {
	var foo Foo
	l, err := net.Listen("tcp", ":9999")
	_ = geerpc.Register(&foo)
	geerpc.HandleHTTP()
	if l == nil {
		log.Println("l nil err ", err)
		return
	}
	addrCh <- l.Addr().String()
	log.Println("startServer addr:", addrCh)

	_ = http.Serve(l, nil)

}

func call(addrCh chan string) {
	client, err := geerpc.DialHTTP("tcp", <-addrCh)

	if client == nil {
		log.Fatal("DialHTTP error ", err)
		return
	}

	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)

	//send request and revice response
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}

			var reply int
			if err := client.Call(context.Background(), "Foo.Aum", args, &reply); err != nil {
				log.Fatal("call Foo.Aum error", err)

			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)

	go call(addr)
	startServer(addr)

}
