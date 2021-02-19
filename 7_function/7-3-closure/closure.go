package main
import "fmt"

func getSequence() func() int  {
	i := 0
	return func () int {
		i += 1
		return i
	} 
}

func main()  {
	nextnum := getSequence()
	fmt.Println(nextnum())
	fmt.Println(nextnum())
	fmt.Println(nextnum())

	nextnum1 := getSequence()
	fmt.Println(nextnum1())
	fmt.Println(nextnum1())
}
