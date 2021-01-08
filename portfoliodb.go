package portfoliodb

import (
	"github.com/docopt/docopt-go"
	"github.com/mitchellh/colorstring"
)

func main() {
	usage := CLIUsage
	args, _ := docopt.ParseDoc(usage)

	if err := dispatchCommand(args); err != nil {
		colorstring.Println("[red][bold]An error occured[reset]")
		colorstring.Println("\t[red]" + err.Error())
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
