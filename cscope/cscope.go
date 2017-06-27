package cscope

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-playground/log"
)

const (
	CSymbol = iota
	FuncDefintion
	FuncsCalledBy
	FuncsCalling
	TextString
	ChangeText
	EgrepPattern
	FindFile
	FilesInclude
)

type Cscope struct {
	binPath, cscopeOutPath string
	SourcePath             string
}

type CscopeResult struct {
	File         string
	FunctionName string
	Line         int
	Text         string
}

var (
	CscopeKind = map[string]int{
		"csymbol":       CSymbol,
		"funcdefintion": FuncDefintion,
		"funcscalledby": FuncsCalledBy,
		"funcscalling":  FuncsCalling,
		"textstring":    TextString,
		"changetext":    ChangeText,
		"egreppattern":  EgrepPattern,
		"findfile":      FindFile,
		"filesinclude":  FilesInclude,
	}
)

func buildCscope(cscopeBinPath, sourcePath string) (out *os.File, err error) {
	log.Info("cscope.files not found, attempting to list C/C++/Java files")

	// check or build for cscope.files
	cscopeFilesPath := path.Join(sourcePath, "cscope.files")
	if _, err = os.Stat(cscopeFilesPath); os.IsNotExist(err) {
		var f *os.File
		if f, err = os.Create(cscopeFilesPath); err != nil {
			return
		}

		// write all files with matching extensions into cscope.files
		err = filepath.Walk(sourcePath, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			switch filepath.Ext(path) {
			case ".cpp", ".cc", ".cxx", ".m", ".hpp", ".hh", ".hxx", ".c", ".h", ".java", ".properties", ".go":
				_, err = f.WriteString(path + "\n")
			default:
				err = nil
			}
			return err
		})

		f.Close()
	}

	log.Info("cscope.out not found, attempting to build index")
	// build the index
	cscopeOutPath := path.Join(sourcePath, "cscope.out")
	cmd := exec.Command(cscopeBinPath, "-b", "-f", cscopeOutPath)
	if _, err = cmd.Output(); err != nil {
		return
	}

	out, err = os.Open(cscopeOutPath)
	return
}

func NewCscope(cscopeBinPath, sourcePath string) (c *Cscope, err error) {
	var bin, out *os.File

	if bin, err = os.Open(cscopeBinPath); err != nil {
		cscopeBinPath, err = exec.LookPath(cscopeBinPath)
		if err != nil {
			return
		}
	}

	defer bin.Close()

	cscopeOutPath := path.Join(sourcePath, "cscope.out")
	if out, err = os.Open(cscopeOutPath); err != nil {
		if out, err = buildCscope(cscopeBinPath, sourcePath); err != nil {
			return
		}
	}

	defer out.Close()

	c = &Cscope{cscopeBinPath, cscopeOutPath, sourcePath}
	return
}

func (c *Cscope) Search(s string, kind int) (results []*CscopeResult, err error) {
	var stdout []byte
	var n int

	k := fmt.Sprintf("-%d", kind)
	cmd := exec.Command(c.binPath, "-dL", "-f", c.cscopeOutPath, k, s)

	if stdout, err = cmd.Output(); err != nil {
		return
	}

	results = make([]*CscopeResult, 0, 10)
	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	for scanner.Scan() {
		s := strings.SplitAfterN(scanner.Text(), " ", 4)

		if len(s) != 4 {
			err = fmt.Errorf("Incorrect cscope result format")
		}

		for i, ss := range s {
			s[i] = strings.Trim(ss, " ")
		}

		n, err = strconv.Atoi(s[2])
		if err != nil {
			return nil, err
		}

		r := &CscopeResult{s[0], s[1], n, s[3]}
		results = append(results, r)
	}

	return
}
