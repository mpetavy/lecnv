package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mpetavy/common"
)

var (
	filemask  *string
	recursive *bool
	dos       *bool
	dry       *bool

	le string
)

func init() {
	common.Init(false, "1.0.0", "", "", "2019", "Line ending converter", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, nil, run, 0)

	filemask = flag.String("f", "", "input file or STDIN")
	recursive = flag.Bool("r", false, "recursive directory search")
	dos = flag.Bool("dos", false, "DOS line ending CRLF")
	dry = flag.Bool("n", false, "Dry run")
}

func convert(path string) error {
	newpath := path + ".new"

	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		common.DebugError(fi.Close())
	}()

	fo, err := os.Create(newpath)
	if err != nil {
		return err
	}
	defer func() {
		common.DebugError(fo.Close())
	}()

	r := bufio.NewReader(fi)

	for {
		line, err := r.ReadString('\n')
		oldlen := len(line)

		if len(line) > 0 {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			newlen := len(line)

			if oldlen != newlen {
				line = line + le
			}

			line = strings.Replace(line, "%", "%%", -1)

			_, err = fmt.Fprintf(fo, line)
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}

	}

	err = fi.Close()
	if err != nil {
		return err
	}
	err = fo.Close()
	if err != nil {
		return err
	}

	fiinfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	foinfo, err := os.Stat(newpath)
	if err != nil {
		return err
	}

	if fiinfo.Size() != foinfo.Size() {
		if *dry {
			fmt.Printf("Would %s\n", path)
		} else {
			fmt.Printf("%s\n", path)

			err := os.Remove(path)
			if err != nil {
				return err
			}

			err = os.Rename(newpath, path)
			if err != nil {
				return err
			}
		}
	}

	common.DebugError(common.FileDelete(newpath))

	return nil
}

func run() error {
	if *dos {
		le = fmt.Sprint("\r\n")
	} else {
		le = fmt.Sprintf("\n")
	}

	fw := common.NewFilewalker(*filemask, *recursive, false, convert)

	return fw.Run()
}

func main() {
	defer common.Done()

	common.Run([]string{"f"})
}
