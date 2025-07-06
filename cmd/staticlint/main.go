package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var mychecks []*analysis.Analyzer

	mychecks = append(mychecks,
		inspect.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		unusedresult.Analyzer,
	)

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	mychecks = append(mychecks, noOsExitAnalyzer)

	multichecker.Main(
		mychecks...,
	)
}
