package main

import "unsafe"

const (
	a = "abcdef"
	b = len(a)
	c = unsafe.Sizeof(a)	
)

const (
	i=1<<iota
	j=3<<iota
	k
	l
)

func main()  {
	println(a, b, c)
	println(i, j, k, l)
}