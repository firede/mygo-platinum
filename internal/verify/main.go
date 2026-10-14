// Command verify runs every check of the repository and reports how many
// passed, as scripts/verify.sh shows them. The checks are go vet, gofmt, a
// build of every package, the exit of go test, every package's result and
// every test function and subtest (snapshots, spec checks, interaction and
// smoke tests), and the tests again at another moment and in another time
// zone, which must pass alike. scripts/checks.txt lists the checks expected: one that does
// not run counts as failed, so that a package that stops building, or a test
// that disappears, lowers the count rather than the total.
//
//	go run ./internal/verify [-write-manifest]
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	manifest = "scripts/checks.txt"
	module   = "github.com/firede/mygo-platinum"
)

type event struct {
	Action  string
	Package string
	Test    string
	Output  string
}

// report is the outcome of the checks: each check passed or failed, and
// what to show about the failures.
type report struct {
	results map[string]bool
	notes   []string
}

func newReport() *report { return &report{results: map[string]bool{}} }

func (r *report) set(name string, ok bool, note string) {
	r.results[name] = ok
	if !ok && note != "" {
		r.notes = append(r.notes, name+":\n"+strings.TrimRight(note, "\n"))
	}
}

// addTests reads the JSON events of go test -json and its exit error into
// checks: one for the run's exit, one for each package, and one for each
// test and subtest — a test with subtests counts for itself as well, so
// that a failure of its own is not lost behind passing subtests.
func (r *report) addTests(stdout []byte, exitErr error, stderr string) {
	r.set("check:go-test-exit", exitErr == nil, fmt.Sprint(exitErr, "\n", stderr))
	outputs := map[string]*strings.Builder{}
	pkgs := map[string]string{}
	tests := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		var ev event
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		pkg := strings.TrimPrefix(ev.Package, module)
		name := "package:" + pkg
		if ev.Test != "" {
			name = "test:" + pkg + ":" + ev.Test
		}
		switch ev.Action {
		case "output":
			if outputs[name] == nil {
				outputs[name] = &strings.Builder{}
			}
			outputs[name].WriteString(ev.Output)
		case "pass", "fail", "skip":
			if ev.Test == "" {
				pkgs[name] = ev.Action
			} else {
				tests[name] = ev.Action
			}
		}
	}
	for name, action := range pkgs {
		note := ""
		if outputs[name] != nil {
			note = outputs[name].String()
		}
		r.set(name, action != "fail", note)
	}
	for name, action := range tests {
		note := ""
		if outputs[name] != nil {
			note = outputs[name].String()
		}
		r.set(name, action != "fail", note)
	}
}

// expect makes every check of the manifest count: one that did not run
// fails. Checks of directories that are not part of the published
// repository are refused.
func (r *report) expect(lines []string) {
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		if strings.Contains(l, ":/local") {
			r.set("check:manifest", false, "the manifest names a check outside the published repository: "+l)
			continue
		}
		if _, ok := r.results[l]; !ok {
			r.set(l, false, "did not run")
		}
	}
}

// count returns the checks passed and run.
func (r *report) count() (passed, total int) {
	for _, ok := range r.results {
		if ok {
			passed++
		}
	}
	return passed, len(r.results)
}

func (r *report) names() []string {
	var names []string
	for n := range r.results {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func run(name string, args ...string) (string, error) {
	return runCmd(exec.Command(name, args...))
}

func runCmd(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return out.String(), err
}

func main() {
	write := flag.Bool("write-manifest", false, "write the checks that ran to "+manifest)
	flag.Parse()
	r := newReport()

	out, err := run("go", "vet", "./...")
	r.set("check:vet", err == nil, out)
	out, err = run("gofmt", "-l", ".")
	r.set("check:gofmt", err == nil && strings.TrimSpace(out) == "", "gofmt -l lists:\n"+out)
	out, err = run("go", "build", "./...")
	r.set("check:build", err == nil, out)

	cmd := exec.Command("go", "test", "-json", "-count=1", "-timeout=240s", "./...")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	r.addTests(stdout, err, stderr.String())

	// The same tests at another moment and in another time zone: a result
	// that depends on either reads the wall clock somewhere it must not.
	cmd = exec.Command("go", "test", "-count=1", "-timeout=240s", "./...")
	cmd.Env = append(os.Environ(), "TZ=Pacific/Kiritimati", "PLATINUM_TEST_NOW=1999-12-31T23:59:59Z")
	out, err = runCmd(cmd)
	r.set("check:deterministic", err == nil, "go test at another moment and time zone:\n"+out)

	var lines []string
	if data, err := os.ReadFile(manifest); err == nil {
		lines = strings.Split(string(data), "\n")
	} else if !*write {
		r.set("check:manifest", false, err.Error())
	}
	inManifest := map[string]bool{}
	for _, l := range lines {
		inManifest[strings.TrimSpace(l)] = true
	}
	r.expect(lines)

	if *write {
		var b strings.Builder
		b.WriteString("# Checks scripts/verify.sh expects to run. Regenerate with\n# go run ./internal/verify -write-manifest after adding tests.\n")
		for _, n := range r.names() {
			if n != "check:manifest" {
				b.WriteString(n + "\n")
			}
		}
		if err := os.WriteFile(manifest, []byte(b.String()), 0o644); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	for _, n := range r.names() {
		if !r.results[n] {
			fmt.Println("FAIL", n)
		} else if !inManifest[n] && !*write {
			fmt.Println("note: not in the manifest:", n)
		}
	}
	for _, n := range r.notes {
		fmt.Println(n)
	}
	passed, total := r.count()
	fmt.Printf("PASSED %d/%d\n", passed, total)
	if passed != total {
		os.Exit(1)
	}
}
