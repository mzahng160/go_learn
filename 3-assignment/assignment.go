package main
var x, y int

var (
	a int
	b bool
)

var c, d int = 1, 2
var e, f = 11, "23_times"

func main(){
	g, h := 12, "assign"
	println(x, y, a, b, c, d, e, f, g, h)
}