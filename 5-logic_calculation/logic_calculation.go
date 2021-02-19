package main

import "fmt"

func main()  {
	a := 8
	b := 3

	c := a & b
	fmt.Printf("a & b %d\n", c)

	c = a | b
	fmt.Printf("a | b %d\n", c)

	d := 7
	e := 4
	c = d ^ e
	fmt.Printf("d ^ e %d\n", c)

	c = a << 2
	fmt.Printf("a << 2 %d\n", c)

	c = a >> 2
	fmt.Printf("a << 2 %d\n", c)

	a <<= 2
	fmt.Printf("a <<= 2 %d\n", a)

	fmt.Printf("&a %d\n", &a)
}