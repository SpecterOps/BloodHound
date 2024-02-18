package golang

import (
	"fmt"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/cmd/vizslas/cve"
	"github.com/specterops/bloodhound/src/cmd/vizslas/golang/vulndb"
	"github.com/specterops/bloodhound/src/cmd/vizslas/ingest"
	"github.com/specterops/bloodhound/src/version"
	"time"
)

type ModuleError struct {
	Err string // the error itself
}

type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

type Module struct {
	Path       string       // module path
	Query      string       // version query corresponding to this version
	Version    string       // module version
	Versions   []string     // available module versions
	Replace    *Module      // replaced by this module
	Time       *time.Time   // time version was created
	Update     *Module      // available update (with -u)
	Main       bool         // is this the main module?
	Indirect   bool         // module is only indirectly needed by main module
	Dir        string       // directory holding local copy of files, if any
	GoMod      string       // path to go.mod file describing module, if any
	GoVersion  string       // go version used in module
	Retracted  []string     // retraction information, if any (with -retracted or -u)
	Deprecated string       // deprecation message, if any (with -u)
	Error      *ModuleError // error loading module
	Origin     any          // provenance of module
	Reuse      bool         // reuse of old module info is safe
}

type Package struct {
	Dir            string   // directory containing package sources
	ImportPath     string   // import path of package in dir
	ImportComment  string   // path in import comment on package statement
	Name           string   // package name
	Doc            string   // package documentation string
	Target         string   // install path
	Shlib          string   // the shared library that contains this package (only set when -linkshared)
	Goroot         bool     // is this package in the Go root?
	Standard       bool     // is this package part of the standard Go library?
	Stale          bool     // would 'go install' do anything for this package?
	StaleReason    string   // explanation for Stale==true
	Root           string   // Go root or Go path dir containing this package
	ConflictDir    string   // this directory shadows Dir in $GOPATH
	BinaryOnly     bool     // binary-only package (no longer supported)
	ForTest        string   // package is only for use in named test
	Export         string   // file containing export data (when using -export)
	BuildID        string   // build ID of the compiled package (when using -export)
	Module         *Module  // info about package's containing module, if any (can be nil)
	Match          []string // command-line patterns matching this package
	DepOnly        bool     // package is only a dependency, not explicitly listed
	DefaultGODEBUG string   // default GODEBUG setting, for main packages

	// Source files
	GoFiles           []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles          []string // .go source files that import "C"
	CompiledGoFiles   []string // .go files presented to compiler (when using -compiled)
	IgnoredGoFiles    []string // .go source files ignored due to build constraints
	IgnoredOtherFiles []string // non-.go source files ignored due to build constraints
	CFiles            []string // .c source files
	CXXFiles          []string // .cc, .cxx and .cpp source files
	MFiles            []string // .m source files
	HFiles            []string // .h, .hh, .hpp and .hxx source files
	FFiles            []string // .f, .F, .for and .f90 Fortran source files
	SFiles            []string // .s source files
	SwigFiles         []string // .swig files
	SwigCXXFiles      []string // .swigcxx files
	SysoFiles         []string // .syso object files to add to archive
	TestGoFiles       []string // _test.go files in package
	XTestGoFiles      []string // _test.go files outside package

	// Embedded files
	EmbedPatterns      []string // //go:embed patterns
	EmbedFiles         []string // files matched by EmbedPatterns
	TestEmbedPatterns  []string // //go:embed patterns in TestGoFiles
	TestEmbedFiles     []string // files matched by TestEmbedPatterns
	XTestEmbedPatterns []string // //go:embed patterns in XTestGoFiles
	XTestEmbedFiles    []string // files matched by XTestEmbedPatterns

	// Cgo directives
	CgoCFLAGS    []string // cgo: flags for C compiler
	CgoCPPFLAGS  []string // cgo: flags for C preprocessor
	CgoCXXFLAGS  []string // cgo: flags for C++ compiler
	CgoFFLAGS    []string // cgo: flags for Fortran compiler
	CgoLDFLAGS   []string // cgo: flags for linker
	CgoPkgConfig []string // cgo: pkg-config names

	// Dependency information
	Imports      []string          // import paths used by this package
	ImportMap    map[string]string // map from source import to ImportPath (identity entries omitted)
	Deps         []string          // all (recursively) imported dependencies
	TestImports  []string          // imports from TestGoFiles
	XTestImports []string          // imports from XTestGoFiles

	// Error information
	Incomplete bool            // this package or a dependency has an error
	Error      *PackageError   // error loading package
	DepsErrors []*PackageError // errors loading dependencies
}

type Workspace struct {
	Modules              []*Module
	Packages             []*Package
	modulesByPath        map[string]int
	packagesByImportPath map[string]int
	cveDB                *cve.Database
	govlunDB             *vulndb.Database
}

func NewWorkspace() *Workspace {
	return &Workspace{
		modulesByPath:        map[string]int{},
		packagesByImportPath: map[string]int{},
	}
}

func (s *Workspace) AddModule(module *Module) {
	s.Modules = append(s.Modules, module)
	s.modulesByPath[module.Path] = len(s.Modules) - 1
}

func (s *Workspace) AddPackage(pkg *Package) {
	s.Packages = append(s.Packages, pkg)

	pkgIdx := len(s.Packages) - 1

	s.packagesByImportPath[pkg.ImportPath] = pkgIdx
}

func (s *Workspace) ModuleByPath(path string) (*Module, bool) {
	if moduleIdx, hasModule := s.modulesByPath[path]; hasModule {
		return s.Modules[moduleIdx], true
	}

	return nil, false
}

func (s *Workspace) PackageByImportPath(importPath string) (*Package, bool) {
	if packageIdx, hasPackage := s.packagesByImportPath[importPath]; hasPackage {
		return s.Packages[packageIdx], true
	}

	return nil, false
}

func (s *Workspace) PackageModule(pkg *Package) (*Module, bool) {
	if module, hasModule := s.ModuleByPath(pkg.ImportPath); hasModule {
		return module, true
	}

	return pkg.Module, pkg.Module != nil
}

func RangeEventAffectsModuleVersion(moduleVersion version.Version, rangeEvent vulndb.RangeEvent) (bool, error) {
	var affected = false

	if rangeEvent.Introduced == "0" {
		affected = true
	} else if affectsVersionsAfter, err := version.Parse(rangeEvent.Introduced); err != nil {
		return false, fmt.Errorf("introduced version is malformed: %v", err)
	} else {
		affected = moduleVersion.Equals(affectsVersionsAfter) && moduleVersion.GreaterThan(affectsVersionsAfter)
	}

	if rangeEvent.Fixed != "" {
		if affectsVersionsBefore, err := version.Parse(rangeEvent.Fixed); err != nil {
			return false, fmt.Errorf("fixed version is malformed: %v", err)
		} else {
			affected = moduleVersion.LessThan(affectsVersionsBefore)
		}
	}

	return affected, nil
}

func LookupVulnCVE(cveDB *cve.Database, vuln vulndb.Entry) (*cve.Entry, bool) {
	for _, alias := range vuln.Aliases {
		if cveEntry, hasCVE := cveDB.LookupCVE(alias); hasCVE {
			return cveEntry, true
		}
	}

	return nil, false
}

func ToIngestPayload(workspace *Workspace) ingest.Payload {
	payload := ingest.Payload{
		Metadata: ingest.Metadata{
			Version: 1,
		},
	}

	entriesByID := make(map[string]int, len(workspace.govlunDB.Entries))
	for idx, entry := range workspace.govlunDB.Entries {
		entriesByID[entry.ID] = idx
	}

	for vulnerableImportPath, vulnerableModule := range workspace.govlunDB.Modules {
		if dependentPkg, hasPkg := workspace.PackageByImportPath(vulnerableImportPath); !hasPkg {
			continue
		} else if dependentPkg.Module.Version == "" {
			log.Errorf("Go module %s has no version", vulnerableImportPath)
		} else if dependentModuleVersion, err := version.Parse(dependentPkg.Module.Version); err != nil {
			log.Errorf("Unable to parse module %s version %s: %v", vulnerableImportPath, dependentPkg.Module.Version, err)
		} else {
			for _, vuln := range vulnerableModule.Vulns {
				if vulnEntryIdx, hasEntry := entriesByID[vuln.ID]; !hasEntry {
					log.Errorf("Unable to find a govuln db entry for vuln %s", vuln.ID)
				} else {
					// TODO: Refactor out taking the pointer for the ingest payload - this function should return its
					//       own payload that can then be merged.
					AnalyzeVulnerableModule(workspace, workspace.govlunDB.Entries[vulnEntryIdx], dependentModuleVersion, &payload, dependentPkg)
				}
			}
		}
	}

	for _, module := range workspace.Modules {
		payload.Visited.AddNode(ingest.Node{
			Entity: ingest.Entity{
				Kind:   "sbom_go_module",
				IDKeys: []string{"path", "version"},
				Properties: map[string]any{
					"path":       module.Path,
					"version":    module.Version,
					"go_version": module.GoVersion,
					"has_update": module.Update != nil,
				},
			},
			//ExtendedKinds: []string{"sbom"},
		})
	}

	for _, pkg := range workspace.Packages {
		payload.Visited.AddNode(ingest.Node{
			Entity: ingest.Entity{
				Kind:   "sbom_go_package",
				IDKeys: []string{"path", "version"},
				Properties: map[string]any{
					"path":       pkg.ImportPath,
					"version":    pkg.Module.Version,
					"go_version": pkg.Module.GoVersion,
					"is_stdlib":  pkg.Standard,
				},
			},
			//ExtendedKinds: []string{"sbom"},
		})

		payload.Visited.AddEdge(ingest.Edge{
			Entity: ingest.Entity{
				Kind: "sbom_contains",
			},

			Start: ingest.Node{
				Entity: ingest.Entity{
					IDKeys: []string{"path", "version"},
					Kind:   "sbom_go_module",
					Properties: map[string]any{
						"path":    pkg.Module.Path,
						"version": pkg.Module.Version,
					},
				},
			},

			End: ingest.Node{
				Entity: ingest.Entity{
					IDKeys: []string{"path", "version"},
					Kind:   "sbom_go_package",
					Properties: map[string]any{
						"path":    pkg.ImportPath,
						"version": pkg.Module.Version,
					},
				},
			},
		})

		for _, pkgImport := range pkg.Imports {
			if dependencyPkg, hasDependencyPkg := workspace.PackageByImportPath(pkgImport); !hasDependencyPkg {
				log.Warnf("No package found for import path: %s", pkgImport)
			} else {
				payload.Visited.AddEdge(ingest.Edge{
					Entity: ingest.Entity{
						Kind: "sbom_depends_on",
					},

					Start: ingest.Node{
						Entity: ingest.Entity{
							IDKeys: []string{"path", "version"},
							Kind:   "sbom_go_package",
							Properties: map[string]any{
								"path":    pkg.ImportPath,
								"version": pkg.Module.Version,
							},
						},
					},

					End: ingest.Node{
						Entity: ingest.Entity{
							IDKeys: []string{"path", "version"},
							Kind:   "sbom_go_package",
							Properties: map[string]any{
								"path":    dependencyPkg.ImportPath,
								"version": dependencyPkg.Module.Version,
							},
						},
					},
				})
			}
		}
	}

	return payload
}

func AnalyzeVulnerableModule(workspace *Workspace, vulnDBEntry vulndb.Entry, dependentModuleVersion version.Version, payload *ingest.Payload, dependentPkg *Package) {
	for _, affectedVersion := range vulnDBEntry.Affected {
		for _, affectedVersionRange := range affectedVersion.Ranges {
			for _, rangeEvent := range affectedVersionRange.Events {
				// Is this an affected version as defined by the range event?
				if affected, err := RangeEventAffectsModuleVersion(dependentModuleVersion, rangeEvent); err != nil {
					log.Errorf("Unable to parse range version information: %v", err)
				} else if !affected {
					// Skip this range event if the module isn't affected.
					continue
				}

				// Create a go vlundb node
				payload.Visited.AddNode(ingest.Node{
					Entity: ingest.Entity{
						Kind:   "sbom_go_vulndb_entry",
						IDKeys: []string{"id"},
						Properties: map[string]any{
							"id":           vulnDBEntry.ID,
							"published_at": vulnDBEntry.Published.Time,
						},
					},
					//ExtendedKinds: []string{"sbom"},
				})

				// Bind the go vulndb entry to the affected module
				payload.Visited.AddEdge(ingest.Edge{
					Entity: ingest.Entity{
						Kind: "sbom_affects",
					},

					Start: ingest.Node{
						Entity: ingest.Entity{
							Kind:   "sbom_go_vulndb_entry",
							IDKeys: []string{"id"},
							Properties: map[string]any{
								"id": vulnDBEntry.ID,
							},
						},
					},

					End: ingest.Node{
						Entity: ingest.Entity{
							Kind:   "sbom_go_module",
							IDKeys: []string{"path", "version"},
							Properties: map[string]any{
								"path":    dependentPkg.Module.Path,
								"version": dependentPkg.Module.Version,
							},
						},
					},
				})

				// Look up the associated CVE
				if vulnCVE, hasCVE := LookupVulnCVE(workspace.cveDB, vulnDBEntry); !hasCVE {
					log.Infof("Unable to find a CVE entry for go vlundb entry: %s", vulnDBEntry.ID)
				} else {
					// Create a CVE node
					payload.Visited.AddNode(ingest.Node{
						Entity: ingest.Entity{
							Kind:   "sbom_cve",
							IDKeys: []string{"id"},
							Properties: map[string]any{
								"id":    vulnCVE.Metadata.ID,
								"title": vulnCVE.Containers.NumberingAuthority.Title,
							},
						},
						//ExtendedKinds: []string{"sbom"},
					})

					// Bind the go vlundb node to the CVE node
					payload.Visited.AddEdge(ingest.Edge{
						Entity: ingest.Entity{
							Kind: "sbom_references",
						},

						Start: ingest.Node{
							Entity: ingest.Entity{
								Kind:   "sbom_go_vulndb_entry",
								IDKeys: []string{"id"},
								Properties: map[string]any{
									"id": vulnDBEntry.ID,
								},
							},
						},

						End: ingest.Node{
							Entity: ingest.Entity{
								Kind:   "sbom_cve",
								IDKeys: []string{"id"},
								Properties: map[string]any{
									"id":    vulnCVE.Metadata.ID,
									"title": vulnCVE.Containers.NumberingAuthority.Title,
								},
							},
						},
					})

					// Stop at the first affected reference
					return
				}
			}
		}
	}
}
