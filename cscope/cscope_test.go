package cscope

import "testing"

func TestCscope(t *testing.T) {
	_, err := NewCscope("cscope", "../fixtures")
	if err != nil {
		t.Error(err)
	}

	_, err = NewCscope("____cscope", "none")
	if err == nil {
		t.Errorf("Expect to not find binary")
	}

	_, err = NewCscope("cscope", "none")
	if err == nil {
		t.Errorf("Expect to not find project")
	}
}

func TestSearch(t *testing.T) {
	cs, err := NewCscope("cscope", "../fixtures")
	if err != nil {
		t.Error(err)
	}

	res, err := cs.Search("listNode", CSymbol)
	if err != nil {
		t.Error(err)
	}

	if res == nil {
		t.Errorf("Expect result to be non-nil")
	}

	t.Logf("%#v", res)
}
