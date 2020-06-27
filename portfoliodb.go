package main

import (
	"fmt"
	
	"github.com/docopt/docopt-go"
)

func main() {
	usage := ReadFile("./USAGE")
	args, _ := docopt.ParseDoc(usage)
	fmt.Println(args)
}
