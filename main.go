package main

import (
	"bufio"
	"bytes"
	"embed"
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

//go:embed go.mod
var resources embed.FS

func init() {
	common.Init("", "", "", "", "Line ending converter", "", "", "", &resources, nil, nil, run, 0)
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

	if *lf && *crlf {
		return fmt.Errorf("either -lf or -crlf, but not both")
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

	var le []byte

	isOnlyList := !*lf && !*crlf
	isCrlf := bytes.Index(ba, []byte{0x0d, 0x0a}) != -1

	switch {
	case *lf:
		if isCrlf {
			isCrlf = false
			le = []byte{0x0a}
		}
	case *crlf:
		if !isCrlf {
			isCrlf = true
			le = []byte{0x0d, 0x0a}
		}
	}

	var info string

	if isCrlf {
		info = "CRLF"
	} else {
		info = "LF"
	}

	if len(le) > 0 {
		info += "*"
	}

	if isOnlyList {
		info = fmt.Sprintf("%-4s", info)
	} else {
		info = fmt.Sprintf("%-5s", info)
	}

	common.Info("%s %s", info, path)

	if isOnlyList || len(le) == 0 {
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
