package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/spf13/cobra"
)

// Each invocation owns its command tree and flag values.
type application struct {
	options options
	data    any
	handled bool
}

func Execute(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	a := &application{}
	root := a.rootCommand()
	root.SetArgs(append([]string{}, args...))
	if stdin == nil {
		stdin = strings.NewReader("")
	}
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(ctx)
	if err == nil && !a.handled {
		// Help and shell completion write text without a JSON trailer.
		return 0
	}
	if err != nil && !a.handled {
		err = clierror.New("usage", err.Error())
	}
	return writeResult(stdout, a.options.output, a.data, err)
}

func (a *application) command(use, short string, run func(*cobra.Command, []string) (any, error)) *cobra.Command {
	return &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			a.handled = true
			var err error
			a.data, err = run(cmd, args)
			return err
		},
	}
}

func commandGroup(use, short string) *cobra.Command {
	return &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
}

func writeResult(stdout io.Writer, output string, data any, err error) int {
	out := map[string]any{"ok": err == nil}
	if data != nil {
		out["data"] = data
	}
	code := 0
	if err != nil {
		e := clierror.From(err)
		out["error"] = e
		switch e.Code {
		case "usage", "validation", "config":
			code = 2
		case "auth", "permission":
			code = 3
		case "conflict", "uncertain":
			code = 4
		case "wait_timeout":
			code = 5
		case "execution":
			code = 6
		default:
			code = 1
		}
	}
	if output == "text" {
		b, _ := json.MarshalIndent(out, "", "  ")
		if _, err := fmt.Fprintln(stdout, string(b)); err != nil {
			return 1
		}
	} else if err := json.NewEncoder(stdout).Encode(out); err != nil {
		return 1
	}
	return code
}
