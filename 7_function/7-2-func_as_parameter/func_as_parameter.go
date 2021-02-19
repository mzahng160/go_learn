package main

import ("fmt")
type cb func(int) int

func main()  {
	testCallback(1, callback)
	testCallback(2, func (x int ) int {
		fmt.Printf("callback2, x:%d\n", x)
		return x
	})
}

func testCallback(x int, f cb)  {
	f(x)
}

func callback(x int) int {
	fmt.Printf("callback, x:%d\n", x)
	return x
}