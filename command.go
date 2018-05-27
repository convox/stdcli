package stdcli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

type Command struct {
	Command     []string
	Description string
	Flags       []Flag
	Handler     HandlerFunc
	Usage       string
	Validate    Validator

	engine *Engine
}

type CommandOptions struct {
	Flags    []Flag
	Usage    string
	Validate Validator
}

type HandlerFunc func(*Context) error

func (c *Command) Execute(args []string) error {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.Usage = func() { helpCommand(c.engine, c) }

	flags := []*Flag{}

	for _, f := range c.Flags {
		g := f
		flags = append(flags, &g)
		flag := fs.VarPF(&g, f.Name, f.Short, f.Description)
		if f.Kind == reflect.Bool {
			flag.NoOptDefVal = "true"
		}
	}

	if err := fs.Parse(args); err != nil {
		if strings.HasPrefix(err.Error(), "unknown shorthand flag") {
			parts := strings.Split(err.Error(), " ")
			return fmt.Errorf("unknown flag: %s", parts[len(parts)-1])
		}
		if err == pflag.ErrHelp {
			return nil
		}
		return err
	}

	ctx := &Context{
		Args:   fs.Args(),
		Flags:  flags,
		engine: c.engine,
	}

	if c.Validate != nil {
		if err := c.Validate(ctx); err != nil {
			return err
		}
	}

	if err := c.Handler(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Command) FullCommand() string {
	return filepath.Base(os.Args[0]) + " " + strings.Join(c.Command, " ")
}

func (c *Command) Match(args []string) ([]string, bool) {
	if len(args) < len(c.Command) {
		return args, false
	}

	for i := range c.Command {
		if args[i] != c.Command[i] {
			return args, false
		}
	}

	return args[len(c.Command):], true
}
