package squirrel

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	symbolTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); !ok {
		color.NoColor = false
	}
}

func TestNonLocalDefinition(t *testing.T) {
	repoDirs, err := os.ReadDir("test_repos")
	fatalIfErrorLabel(t, err, "reading test_repos")

	annotations := []annotation{}

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		contents, err := os.ReadFile(filepath.Join("test_repos", path.Repo, path.Path))
		fatalIfErrorLabel(t, err, "reading a file")
		return contents, nil
	}

	tempSquirrel := New(readFile, nil)
	allSymbols := []result.Symbol{}

	for _, repoDir := range repoDirs {
		if !repoDir.IsDir() {
			t.Fatalf("unexpected file %s", repoDir.Name())
		}

		base := filepath.Join("test_repos", repoDir.Name())
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			contents, err := os.ReadFile(path)
			fatalIfErrorLabel(t, err, "reading annotations from a file")

			rel, err := filepath.Rel(base, path)
			fatalIfErrorLabel(t, err, "getting relative path")
			repoCommitPath := types.RepoCommitPath{Repo: repoDir.Name(), Commit: "abc", Path: rel}

			annotations = append(annotations, collectAnnotations(repoCommitPath, string(contents))...)

			symbols, err := tempSquirrel.getSymbols(context.Background(), repoCommitPath)
			fatalIfErrorLabel(t, err, "getSymbols")
			allSymbols = append(allSymbols, symbols...)

			return nil
		})
		fatalIfErrorLabel(t, err, "walking a repo dir")
	}

	ss := func(ctx context.Context, args symbolTypes.SearchArgs) (result.Symbols, error) {
		results := result.Symbols{}
	nextSymbol:
		for _, s := range allSymbols {
			if args.IncludePatterns != nil {
				for _, p := range args.IncludePatterns {
					match, err := regexp.MatchString(p, s.Path)
					fatalIfErrorLabel(t, err, "matching a pattern")
					if !match {
						continue nextSymbol
					}
				}
			}
			match, err := regexp.MatchString(args.Query, s.Name)
			if err != nil {
				return nil, err
			}
			if match {
				results = append(results, s)
			}
		}
		return results, nil
	}

	squirrel := New(readFile, ss)
	squirrel.errorOnParseFailure = true
	defer squirrel.Close()

	cwd, err := os.Getwd()
	fatalIfErrorLabel(t, err, "getting cwd")

	symbolToKindToAnnotations := groupBySymbolAndKind(annotations)
	symbols := []string{}
	for symbol := range symbolToKindToAnnotations {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		m := symbolToKindToAnnotations[symbol]
		var wantAnn *annotation
		for _, ann := range m["def"] {
			if wantAnn != nil {
				t.Fatalf("multiple definitions for symbol %s", symbol)
			}

			annCopy := ann
			wantAnn = &annCopy
		}

		if wantAnn == nil {
			t.Fatalf("no matching \"def\" annotation for \"ref\" %s", symbol)
		}

		want := wantAnn.repoCommitPathPoint

		for _, ref := range m["ref"] {
			squirrel.breadcrumbs = Breadcrumbs{}
			gotSymbolInfo, err := squirrel.symbolInfo(context.Background(), ref.repoCommitPathPoint)
			fatalIfErrorLabel(t, err, "symbolInfo")

			if gotSymbolInfo == nil {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
				t.Fatalf("no symbolInfo for symbol %s", symbol)
			}

			got := types.RepoCommitPathPoint{
				RepoCommitPath: gotSymbolInfo.Definition.RepoCommitPath,
				Point: types.Point{
					Row:    gotSymbolInfo.Definition.Row,
					Column: gotSymbolInfo.Definition.Column,
				},
			}

			if diff := cmp.Diff(wantAnn.repoCommitPathPoint, got); diff != "" {
				t.Errorf("wrong symbolInfo for %q\n", symbol)
				t.Errorf("want: %s%s/%s:%d:%d\n", itermSource(filepath.Join(cwd, "test_repos", want.Repo, want.Path), want.Point.Row, "src"), want.Repo, want.Path, want.Point.Row, want.Point.Column)
				t.Errorf("got : %s%s/%s:%d:%d\n", itermSource(filepath.Join(cwd, "test_repos", got.Repo, got.Path), got.Point.Row, "src"), got.Repo, got.Path, got.Point.Row, got.Point.Column)
			}
		}
	}
}

func groupBySymbolAndKind(annotations []annotation) map[string]map[string][]annotation {
	grouped := map[string]map[string][]annotation{}

	for _, a := range annotations {
		if _, ok := grouped[a.symbol]; !ok {
			grouped[a.symbol] = map[string][]annotation{}
		}

		if _, ok := grouped[a.symbol][a.kind]; !ok {
			grouped[a.symbol][a.kind] = []annotation{}
		}

		grouped[a.symbol][a.kind] = append(grouped[a.symbol][a.kind], a)
	}

	return grouped
}
