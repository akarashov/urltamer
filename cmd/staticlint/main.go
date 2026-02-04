// Multichecker static analyzer.
//   - standard static analyzers golang.org/x/tools/go/analysis/passes
//   - all analyzers class SA from staticcheck honnef.co/go/tools/staticcheck
//   - one more other analyzers from staticcheck honnef.co/go/tools/staticcheck
//   - two public analyzers: `github.com/mdempsky/unconvert` and `github.com/mvdan/unparam`.
//   - a custom analyzer `NoExitMain` which forbids calling `os.Exit` inside the `main` function.
//
// # Running
//
//	go run ./cmd/staticlint ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Standard static analyzers.
	analyzers = append(analyzers,
		asmdecl.Analyzer,       // Check assembly declarations.
		assign.Analyzer,        // Check for useless assignments.
		atomic.Analyzer,        // Check atomic package usage.
		bools.Analyzer,         // Check boolean operators.
		buildtag.Analyzer,      // Check build tags.
		cgocall.Analyzer,       // Check cgo calls.
		composite.Analyzer,     // Check composite literals.
		copylock.Analyzer,      // Check for copied locks.
		errorsas.Analyzer,      // Check errors.As usage.
		httpresponse.Analyzer,  // Check HTTP response usage.
		loopclosure.Analyzer,   // Check loop closures.
		lostcancel.Analyzer,    // Check for lost context cancellation.
		nilfunc.Analyzer,       // Check nil function comparisons.
		printf.Analyzer,        // Check printf format strings.
		shadow.Analyzer,        // Check for variable shadowing.
		shift.Analyzer,         // Check bit shifts.
		sortslice.Analyzer,     // Check sort.Slice usage.
		stdmethods.Analyzer,    // Check standard method signatures.
		stringintconv.Analyzer, // Check string/int conversions.
		structtag.Analyzer,     // Check struct tags.
		tests.Analyzer,         // Check test functions.
		unmarshal.Analyzer,     // Check unmarshal usage.
		unreachable.Analyzer,   // Check for unreachable code.
		unsafeptr.Analyzer,     // Check unsafe pointer usage.
		unusedresult.Analyzer,  // Check for unused results.
	)

	// All analyzers class SA from staticcheck
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// One more other analyzers from staticcheck (Simplification)
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// External analyzer

	// Our custom analyzer
	analyzers = append(analyzers, NoExitMain)

	multichecker.Main(analyzers...)
}
