package main

import (
	"context"
	"fmt"
	"geerpc"
	"geerpc/xclient"
	"log"
	"net"
	"sync"
	"time"
)

type Foo int
type Args struct{ Num1, Num2 int }

func (f Foo) Aum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addrCh chan string) {
	var foo Foo
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("startServer Listen err:", err)
		return
	}

	server := geerpc.NewServer()
	_ = server.Register(&foo)
	addrCh <- l.Addr().String()
	log.Println("startServer addr:", addrCh)

	server.Accept(l)
}

func glog(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error

	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)

	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}

	log.Printf("#######################error: %v", err)

	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiSercerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)

	defer func() { _ = xc.Close() }()

	//send request and receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			glog(xc, context.Background(), "call", "Foo.Aum", &Args{Num1: i, Num2: i + 1})
		}(i)
	}
	wg.Wait()
}

/*
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
*/
func broadcast(addr1, addr2 string) {

	fmt.Println("broadcast start")

	d := xclient.NewMultiSercerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)

	defer func() { _ = xc.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {

			fmt.Println("broadcast deal:", i)

			defer wg.Done()
			//glog(xc, context.Background(), "broadcast", "Foo.Aum", &Args{Num1: i, Num2: i + 2})

			ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
			glog(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i + 3})

		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)

	//start two server
	go startServer(ch1)
	go startServer(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	//call(addr1, addr2)

	broadcast(addr1, addr2)
}
