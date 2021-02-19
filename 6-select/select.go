package main

import (
	"fmt"
	"time"
)

func Channel(ch chan int, stopCh chan bool){
	i := 10
	for j := 0; j < 10; j++{
		ch <- i
		time.Sleep(time.Second)
	}
	stopCh <- true
}

func main()  {
	ch := make(chan int)
	c := 0
	stopCh := make(chan bool)

	go Channel(ch, stopCh)

	for{
		select {
		case c = <- ch:
			fmt.Println("receive", c)
			fmt.Println("channel")
		case s := <- ch:
			fmt.Println("Receive", s)
		case _ = <- stopCh:
			goto end
		}
	}
	end:
}

