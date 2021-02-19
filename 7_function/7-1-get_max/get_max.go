package main

import "fmt"

func max(num1, num2 int) int {
	//var result int;

	if(num1 > num2){
		return num1
	} else {
		return num2
	}
}

func swap(x, y string) (string, string) {
	return y, x
}

func main()  {
	var a int = 100
	var b int = 200
	var ret int

	ret = max(a, b)
	fmt.Printf("max is %d\n", ret)

	stra, strb := swap("aaa","bbb")
	fmt.Printf("%s, %s\n", stra, strb)
}