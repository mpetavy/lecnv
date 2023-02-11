package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	filemask  = flag.String("f", "", "input file or STDIN")
	recursive = flag.Bool("r", false, "recursive directory search")
	lf        = flag.Bool("lf", false, "set file ending to LF (*nix)")
	crlf      = flag.Bool("crlf", false, "set file ending to CRLF (Windows)")
)

func init() {
	common.Init("1.0.0", "", "", "2019", "Line ending converter", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, nil, run, 0)
}

var (
	exts = []string{
		"*.go",
		"go.mod",
		"go.sum",
		"*.md",
		"*.i18n",
		".xml",
		".sh",
		".js",
		".yml",
		"*.txt",
		".dockerignore",
		".gitignore",
		"LICENSE",
	}
)

func needsProcess(path string) bool {
	if common.ContainsWildcard(*filemask) {
		return true
	}

	for _, ext := range exts {
		b, err := common.EqualWildcards(filepath.Base(path), ext)
		common.Panic(err)

		if b {
			return true
		}
	}

	return false
}

func run() error {
	fw, err := common.NewFilewalker(*filemask, *recursive, true, func(path string, fi os.FileInfo) error {
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") {
				return fs.SkipDir
			} else {
				return nil
			}
		}

		if !needsProcess(path) {
			return nil
		}

		return processFile(path)
	})
	if common.Error(err) {
		return err
	}

	err = fw.Run()
	if common.Error(err) {
		return err
	}

	return nil
}

func processFile(path string) error {
	ba, err := os.ReadFile(path)
	if common.Error(err) {
		return err
	}

	if bytes.Index(ba, []byte{0x0d, 0x0a}) != -1 {
		common.Info("CRLF %s", path)
	} else {
		common.Info("LF %s", path)
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
