// Command benchreport renders the benchmark report of go-json from the results of the machines: the HTML page,
// the SVG summary which the README shows, and a copy of every result file, which the page links to.
//
//	go run ./cmd/benchreport -out site results-amd64.json results-arm64.json
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"benchmark/report"
)

func main() {
	out := flag.String("out", "site", "the directory to write the site to")
	attestations := flag.String("attestations", "", "the URL of the attestations of the result files")
	flag.Parse()
	if err := run(*out, *attestations, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out, attestations string, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no result file is given")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	site := &report.Site{Attestations: attestations}
	for _, f := range files {
		r, err := report.ReadRun(f)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		name := filepath.Base(f)
		// the file is copied as it is, so that its digest is the one of the attested file
		if err := os.WriteFile(filepath.Join(out, name), data, 0o644); err != nil {
			return err
		}
		site.Runs = append(site.Runs, r)
		site.Files = append(site.Files, name)
	}
	page, err := site.HTML()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "index.html"), []byte(page), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "summary.svg"), []byte(site.Summary()), 0o644)
}
