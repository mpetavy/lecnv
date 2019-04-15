package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpetavy/common"
)

var (
	path      *string
	recursive *bool
	dos       *bool
	le        string
)

func init() {
	path = flag.String("p", "", "path to directory")
	recursive = flag.Bool("r", false, "recursive directory walk")
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

func scan(path string) error {
	path = common.CleanPath(path)
	if common.IsFile(path) {
		return convert(path)
	}

	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if !strings.HasPrefix(fi.Name(), ".") {
			fn := filepath.Join(path, fi.Name())

			if common.IsFile(fn) {
				ext := filepath.Ext(fn)
				if ext == ".java" || ext == ".xml" {
					err := convert(fn)
					if err != nil {
						return err
					}
				}
			} else {
				if *recursive {
					err := scan(fn)
					if err != nil {
						return err
					}
				}
			}
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

	b, err := common.FileExists(*path)

	if err != nil {
		return err
	}

	if !b {
		return fmt.Errorf("unknown file or directory: %s", *path)
	}

	err = scan(*path)

	return err
}

func main() {
	defer common.Cleanup()

	common.New(&common.App{"lecnv", "1.0.0", "2019", "Line ending converter", "mpetavy", common.APACHE, "https://github.com/mpetavy/lecnv", false, nil,nil, nil, run, 0}, []string{"p"})
	common.Run()
}
