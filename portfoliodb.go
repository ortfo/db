package main

import (
	"github.com/docopt/docopt-go"
)

func main() {
	usage := ReadFile("./USAGE")
	args, _ := docopt.ParseDoc(usage)
	
	if err := dispatchCommand(args); err != nil {
		panic(err)
	}
}

func dispatchCommand(args docopt.Opts) error {
	if val, _ := args.Bool("build"); val {
		err := RunCommandBuild(args)
		return err
	}
	if val, _ := args.Bool("replicate"); val {
		err := RunCommandReplicate(args)
		return err
	}
	if val, _ := args.Bool("add"); val {
		err := RunCommandAdd(args)
		return err
	}
	if val, _ := args.Bool("validate"); val {
		err := RunCommandValidate(args)
		return err
	}
	return nil
}
