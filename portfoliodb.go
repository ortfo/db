package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/docopt/docopt-go"
)

func main() {
	usage := ReadFile("./USAGE")
	args, _ := docopt.ParseDoc(usage)
	dispatchCommand(args)
	conf, err := GetConfiguration(".")
	if err != nil {
		panic(err)
	}
	spew.Dump(conf)
}

func dispatchCommand(args docopt.Opts) {
	if val, _ := args.Bool("build"); val {
		RunCommandBuild(args)
		return
	}
	if val, _ := args.Bool("replicate"); val {
		RunCommandReplicate(args)
		return
	}
	if val, _ := args.Bool("add"); val {
		RunCommandAdd(args)
		return
	}
	if val, _ := args.Bool("validate"); val {
		RunCommandValidate(args)
		return
	}
}
