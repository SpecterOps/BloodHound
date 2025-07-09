package complexity

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/packages/go/genericgraph"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const (
	Name  = "complexity"
	Usage = "Create generic graph output for combined call and cyclomatic graphs"
)

type command struct {
	env            environment.Environment
	tags           string
	additionalArgs []string
}

// Create new instance of command to capture given environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage of command
func (s *command) Usage() string {
	return Usage
}

// Name of command
func (s *command) Name() string {
	return Name
}

// Parse command flags
func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)

	tags := cmd.String("tags", "", "additional Go build tags")
	s.tags = *tags

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	s.additionalArgs = cmd.Args()

	return nil
}

// Run complexity command
func (s *command) Run() error {
	var (
		cfg = &packages.Config{
			Mode:       packages.LoadAllSyntax,
			BuildFlags: []string{"-tags=" + s.tags},
		}
		outGraph     = genericgraph.GenericObject{}
		nodeMap      = make(map[string]genericgraph.Node)
		fnStringToID = make(map[string]string)
	)

	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	}

	modFile, err := os.ReadFile(filepath.Join(paths.GoModules[0], "go.mod"))
	if err != nil {
		return fmt.Errorf("reading first go.mod file for workspace: %w", err)
	}

	currentModule := modfile.ModulePath(modFile)

	slog.Info("current module", slog.String("module", currentModule))

	initial, err := packages.Load(cfg, s.additionalArgs...)
	if err != nil {
		return err
	}

	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	mode := ssa.InstantiateGenerics // instantiate generics by default for soundness
	prog, _ := ssautil.AllPackages(initial, mode)
	prog.Build()

	// results := rta.Analyze(roots, true)
	callgraph := vta.CallGraph(ssautil.AllFunctions(prog), nil)

	for fn, node := range callgraph.Nodes {
		var (
			pkg      string
			id       = strconv.Itoa(node.ID)
			position = fn.Prog.Fset.PositionFor(fn.Pos(), false).String()
		)

		if fn.Package() != nil {
			pkg = fn.Package().Pkg.String()
		}

		if !strings.Contains(pkg, currentModule) {
			continue
		}

		if _, ok := nodeMap[id]; !ok {
			fnStringToID[fn.String()] = id

			n := genericgraph.Node{
				ID:    id,
				Kinds: []string{"Function", "Golang"},
				Properties: map[string]any{
					"name":     fn.Name(),
					"position": position,
					"pkg":      pkg,
					"relname":  fn.String(),
				},
			}

			nodeMap[id] = n

			outGraph.Graph.Nodes = append(outGraph.Graph.Nodes, n)
		}

		for _, edge := range node.Out {
			var (
				calleePkg string
				calleeID  = strconv.Itoa(edge.Callee.ID)
			)

			if edge.Callee.Func.Package() != nil {
				calleePkg = edge.Callee.Func.Package().Pkg.String()
			}

			if !strings.Contains(calleePkg, currentModule) {
				continue
			}

			outGraph.Graph.Edges = append(outGraph.Graph.Edges, genericgraph.Edge{
				Start: genericgraph.Terminal{
					MatchBy: "id",
					Value:   id,
				},
				End: genericgraph.Terminal{
					MatchBy: "id",
					Value:   calleeID,
				},
				Kind: "Calls",
			})
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err = enc.Encode(outGraph)
	if err != nil {
		return fmt.Errorf("marshaling output graph: %w", err)
	}

	return nil
}
