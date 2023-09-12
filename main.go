package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"io/fs"
	"os"
	"strings"
)

var (
	filemask  = flag.String("f", "", "input file or STDIN")
	recursive = flag.Bool("r", false, "recursive directory search")
	lf        = flag.Bool("lf", false, "set file ending to LF (*nix)")
	crlf      = flag.Bool("crlf", false, "set file ending to CRLF (Windows)")
)

func init() {
	common.Init("lecnv", "", "", "", "2019", "Line ending converter", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, nil, run, 0)
}

func run() error {
	err := common.WalkFiles(*filemask, *recursive, true, func(path string, fi os.FileInfo) error {
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") {
				return fs.SkipDir
			} else {
				return nil
			}
		}

		return processFile(path)
	})
	if common.Error(err) {
		return err
	}

	return nil
}

func processFile(path string) error {
	l, err := common.FileSize(path)
	if common.Error(err) {
		return err
	}

	l = min(int64(4096), l)

	ba := make([]byte, int(l))

	f, err := os.Open(path)
	if common.Error(err) {
		return err
	}

	defer func() {
		common.Error(f.Close())
	}()

	n, err := f.Read(ba)
	if common.Error(err) {
		return err
	}

	ba = ba[:n]

	if bytes.Index(ba, []byte{0x0d, 0x0a}) != -1 {
		common.Info("CRLF %s", path)
	} else {
		common.Info("LF   %s", path)
	}

	var le []byte

	switch {
	case *lf:
		le = []byte{0x0a}
	case *crlf:
		le = []byte{0x0d, 0x0a}
	default:
		return nil
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(ba))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = common.FileBackup(path)
	if common.Error(err) {
		return err
	}

	sb := bytes.Buffer{}

	for _, line := range lines {
		sb.WriteString(line)
		sb.Write(le)
	}

	err = os.WriteFile(path, sb.Bytes(), common.DefaultFileMode)
	if common.Error(err) {
		return err
	}

	return nil
}

func main() {
	common.Run(nil)
}
