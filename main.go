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

	le string
)

func init() {
	filemask = flag.String("f", "", "input file or STDIN")
	recursive = flag.Bool("r", false, "recursive directory search")
	dos = flag.Bool("dos", common.IsWindowsOS(), "DOS line ending CRLF")
}

func convert(path string) error {
	newpath := path + ".new"

	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()

	fo, err := os.Create(newpath)
	if err != nil {
		return err
	}
	defer fo.Close()

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
		fmt.Printf("%s\n", path)

		err := os.Remove(path)
		if err != nil {
			return err
		}

		err = os.Rename(newpath, path)
		if err != nil {
			return err
		}
	} else {
		err := os.Remove(newpath)
		if err != nil {
			return err
		}
	}

	return nil
}

func run() error {
	if *dos {
		le = fmt.Sprint("\r\n")
	} else {
		le = fmt.Sprintf("\n")
	}

	return common.WalkFilepath(*filemask, *recursive, convert)
}

func main() {
	defer common.Cleanup()

	common.New(&common.App{"lecnv", "1.0.0", "2019", "Line ending converter", "mpetavy", common.APACHE, "https://github.com/mpetavy/lecnv", false, nil, nil, nil, run, 0}, []string{"f"})
	common.Run()
}
