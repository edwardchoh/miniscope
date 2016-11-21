package cscope

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	CSymbol = iota
	FuncDefintion
	FuncsCalledBy
	FuncsCalling
	TextString
	EgrepPattern
	FilesInclude
)

type Cscope struct {
	bin, out string
}

type CscopeResult struct {
	File         string
	FunctionName string
	Line         int
	Text         string
}

func NewCscope(cscopeBinPath, cscopeOutPath string) (c *Cscope, err error) {
	var bin, out *os.File

	if bin, err = os.Open(cscopeBinPath); err != nil {
		cscopeBinPath, err = exec.LookPath(cscopeBinPath)
		if err != nil {
			return
		}
	}

	defer bin.Close()

	if out, err = os.Open(cscopeOutPath); err != nil {
		return
	}

	defer out.Close()

	c = &Cscope{cscopeBinPath, cscopeOutPath}
	return
}

func (c *Cscope) Search(s string, kind int) (results []*CscopeResult, err error) {
	var stdout []byte
	var n int

	search := fmt.Sprintf("-%d%s", kind, s)
	cmd := exec.Command(c.bin, "-dL", "-f", c.out, search)

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

		n, err = strconv.Atoi(strings.Trim(s[2], " "))
		if err != nil {
			return nil, err
		}

		r := &CscopeResult{s[0], s[1], n, s[3]}
		results = append(results, r)
	}

	return
}
