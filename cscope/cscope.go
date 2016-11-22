package cscope

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
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
		return
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
