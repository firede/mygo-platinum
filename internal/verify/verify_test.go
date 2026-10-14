package main

import (
	"errors"
	"strings"
	"testing"
)

const pkg = module + "/examples/x"

func events(lines ...string) []byte { return []byte(strings.Join(lines, "\n")) }

// TestExitFailureCounts is the case of a test binary whose TestMain exits
// non-zero after every test passed: the exit and the package fail.
func TestExitFailureCounts(t *testing.T) {
	r := newReport()
	r.addTests(events(
		`{"Action":"run","Package":"`+pkg+`","Test":"TestA"}`,
		`{"Action":"pass","Package":"`+pkg+`","Test":"TestA"}`,
		`{"Action":"output","Package":"`+pkg+`","Output":"FAIL\n"}`,
		`{"Action":"fail","Package":"`+pkg+`"}`,
	), errors.New("exit status 1"), "")
	if r.results["check:go-test-exit"] {
		t.Error("a failing go test exit counted as passed")
	}
	if r.results["package:/examples/x"] {
		t.Error("a failing package counted as passed")
	}
	if !r.results["test:/examples/x:TestA"] {
		t.Error("the passing test counted as failed")
	}
	if p, n := r.count(); p != 1 || n != 3 {
		t.Errorf("counted %d/%d, want 1/3", p, n)
	}
}

// TestParentFailureCounts is a test that fails on its own while its
// subtests pass.
func TestParentFailureCounts(t *testing.T) {
	r := newReport()
	r.addTests(events(
		`{"Action":"pass","Package":"`+pkg+`","Test":"TestP/sub"}`,
		`{"Action":"fail","Package":"`+pkg+`","Test":"TestP"}`,
		`{"Action":"fail","Package":"`+pkg+`"}`,
	), errors.New("exit status 1"), "")
	if r.results["test:/examples/x:TestP"] {
		t.Error("the failing parent counted as passed")
	}
	if !r.results["test:/examples/x:TestP/sub"] {
		t.Error("the passing subtest counted as failed")
	}
}

// TestMissingChecksFail is a manifest naming checks that did not run, and
// one outside the published repository.
func TestMissingChecksFail(t *testing.T) {
	r := newReport()
	r.addTests(events(`{"Action":"pass","Package":"`+pkg+`","Test":"TestA"}`, `{"Action":"pass","Package":"`+pkg+`"}`), nil, "")
	r.expect([]string{"# a comment", "test:/examples/x:TestA", "test:/examples/x:TestGone", "test:/local/exp:TestRender"})
	if ok, ran := r.results["test:/examples/x:TestGone"]; !ran || ok {
		t.Error("a check that did not run is not a failure")
	}
	if r.results["check:manifest"] {
		t.Error("a manifest naming local/ was accepted")
	}
	if p, n := r.count(); p != 3 || n != 5 {
		t.Errorf("counted %d/%d, want 3/5", p, n)
	}
}
