package complexity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/genericgraph"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/cfg"
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
		conf = &packages.Config{
			Mode:       packages.LoadAllSyntax,
			BuildFlags: []string{"-tags=" + s.tags},
		}
		outGraph             = genericgraph.GenericObject{}
		nodeMap              = make(map[string]genericgraph.Node)
		ssaIDToUUID          = make(map[int]uuid.UUID)
		fnPosToUUID          = make(map[string]uuid.UUID)
		declsForFile         = make(map[string][]ast.Decl)
		hashNameToHash       = make(map[string]string)
		calleeHashNameToHash = make(map[string]string)
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

	initial, err := packages.Load(conf, s.additionalArgs...)
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
			pkg         string
			position    = prog.Fset.Position(fn.Pos())
			positionStr = position.String()
		)

		if fn.Package() != nil {
			pkg = fn.Package().Pkg.String()
		}

		if fn.Synthetic != "" {
			continue
		}

		if !strings.Contains(pkg, currentModule) {
			continue
		}

		fnID, ok := ssaIDToUUID[node.ID]
		if !ok {
			fnID, _ = uuid.NewV4()
			ssaIDToUUID[node.ID] = fnID
		}

		_, ok = fnPosToUUID[positionStr]
		if !ok {
			fnPosToUUID[positionStr] = fnID
		}

		hashName := positionStr + fn.Name()
		sha := sha256.Sum256([]byte(hashName))
		id := hex.EncodeToString(sha[:])

		hashNameToHash[hashName] = id

		n := genericgraph.Node{
			ID:    id,
			Kinds: []string{"Function", "Golang"},
			Properties: map[string]any{
				"name":     fn.Name(),
				"position": positionStr,
				"pkg":      pkg,
				"relname":  fn.String(),
			},
		}

		nodeMap[id] = n

		outGraph.Graph.Nodes = append(outGraph.Graph.Nodes, n)

		for _, edge := range node.Out {
			var (
				calleePkg         string
				calleePosition    = prog.Fset.Position(edge.Callee.Func.Pos())
				calleePositionStr = calleePosition.String()
			)

			if edge.Callee.Func.Package() != nil {
				calleePkg = edge.Callee.Func.Package().Pkg.String()
			}

			if !strings.Contains(calleePkg, currentModule) {
				continue
			}

			if edge.Callee.Func.Synthetic != "" {
				continue
			}

			calleeFnID, ok := ssaIDToUUID[edge.Callee.ID]
			if !ok {
				calleeFnID, _ = uuid.NewV4()
				ssaIDToUUID[edge.Callee.ID] = calleeFnID
			}

			_, ok = fnPosToUUID[calleePositionStr]
			if !ok {
				fnPosToUUID[calleePositionStr] = calleeFnID
			}

			calleeHashName := calleePositionStr + edge.Callee.Func.Name()
			calleeSHA := sha256.Sum256([]byte(calleeHashName))
			calleeID := hex.EncodeToString(calleeSHA[:])

			calleeHashNameToHash[calleeHashName] = calleeID

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

	prog.Fset.Iterate(func(file *token.File) bool {
		fset := token.NewFileSet()
		if !strings.Contains(file.Name(), paths.GoModules[0]) {
			return true
		}

		slog.Info("Processing File", slog.String("filename", file.Name()))

		decls, ok := declsForFile[file.Name()]
		if !ok {
			astFile, err := parser.ParseFile(fset, file.Name(), nil, parser.ParseComments)
			if err != nil {
				return true
			}
			decls = astFile.Decls
			declsForFile[file.Name()] = decls
		}

		for _, d := range decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				hashName := fset.Position(decl.Name.Pos()).String() + decl.Name.Name
				sha := sha256.Sum256([]byte(hashName))
				id := hex.EncodeToString(sha[:])
				if decl.Body != nil {
					addControlFlowToGraph(id, cfg.New(decl.Body, func(*ast.CallExpr) bool { return true }), &outGraph.Graph)
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, value := range valueSpec.Values {
						funcLit, ok := value.(*ast.FuncLit)
						if !ok {
							continue
						}
						if funcLit.Body != nil {
							hashName := fset.Position(funcLit.Pos()).String() + fset.Position(funcLit.End()).String()
							sha := sha256.Sum256([]byte(hashName))
							id := hex.EncodeToString(sha[:])
							addControlFlowToGraph(id, cfg.New(funcLit.Body, func(*ast.CallExpr) bool { return true }), &outGraph.Graph)
						}
					}
				}
			}
		}
		return true
	})

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err = enc.Encode(outGraph)
	if err != nil {
		return fmt.Errorf("marshaling output graph: %w", err)
	}

	return nil
}

func addControlFlowToGraph(fnID string, c *cfg.CFG, graph *genericgraph.Graph) {
	for itr, block := range c.Blocks {
		kind := block.Kind.String()

		id := fnID + strconv.Itoa(int(block.Index))

		graph.Nodes = append(graph.Nodes, genericgraph.Node{
			ID:    id,
			Kinds: []string{"ControlFlow", "Golang"},
			Properties: map[string]any{
				"name": kind,
			},
		})

		for _, succ := range block.Succs {
			// Use the function node itself in place of the body node, but that means we need to attach all relationships that
			// would have gone to the Body node to instead point directly at the Function node.
			var startID string
			if itr == 0 {
				startID = fnID
			} else {
				startID = id
			}

			graph.Edges = append(graph.Edges, genericgraph.Edge{
				Start: genericgraph.Terminal{
					MatchBy: "id",
					Value:   startID,
				},
				End: genericgraph.Terminal{
					MatchBy: "id",
					Value:   fnID + strconv.Itoa(int(succ.Index)),
				},
				Kind: "FlowsInto",
			})
		}
	}
}
