package cli

import (
	"io"
	"os"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
)

func readInput(file string, stdin io.Reader, v any) error {
	if file == "" {
		return clierror.New("usage", "需要 --file FILE 或 --file -")
	}
	var r io.Reader
	var f *os.File
	if file == "-" {
		r = stdin
	} else {
		var err error
		f, err = os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	b, err := io.ReadAll(io.LimitReader(r, 10<<20+1))
	if err != nil {
		return err
	}
	if len(b) > 10<<20 {
		return clierror.New("validation", "输入超过 10MiB")
	}
	return jsonutil.Decode(b, v)
}
