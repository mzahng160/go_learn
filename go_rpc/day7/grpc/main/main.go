package main

import (
	"context"
	"fmt"
	"geerpc"
	"geerpc/registry"
	"geerpc/xclient"
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

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

/* func startServer(addrCh chan string) {
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
} */

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := geerpc.NewServer()
	_ = server.Register(&foo)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
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

func call(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
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
func broadcast(registry string) {

	fmt.Println("broadcast start")

	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)

	defer func() { _ = xc.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {

			fmt.Println("broadcast deal:", i)

			defer wg.Done()
			glog(xc, context.Background(), "broadcast", "Foo.Aum", &Args{Num1: i, Num2: i + 2})

			ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
			glog(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i + 3})

		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)

	//start two server
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	//broadcast(registryAddr)
}
