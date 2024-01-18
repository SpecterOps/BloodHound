package analysis

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/slices"
)

const (
	Name  = "analysis"
	Usage = "Run static analyzers"
)

type esLintEntry struct {
	FilePath        string          `json:"filePath"`
	ErrorCount      int             `json:"errorCount"`
	WarningCount    int             `json:"warningCount"`
	FatalErrorCount int             `json:"fatalErrorCount"`
	Messages        []esLintMessage `json:"messages"`
}

type esLintMessage struct {
	RuleID   string `json:"ruleId"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Line     uint64 `json:"line"`
}

type codeClimateEntry struct {
	Description string              `json:"description"`
	Severity    string              `json:"severity"`
	Location    codeClimateLocation `json:"location"`
}

type codeClimateLocation struct {
	Path  string           `json:"path"`
	Lines codeClimateLines `json:"lines"`
}

type codeClimateLines struct {
	Begin uint64 `json:"begin"`
}

type Config struct {
	Environment []string
}

type command struct {
	config Config
}

func (s command) Usage() string {
	return Usage
}

func (s command) Name() string {
	return Name
}

func (s command) Run() error {
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("could not find workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse module absolute paths: %w", err)
	} else if jsPaths, err := workspace.ParseJsAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse JS absolute paths: %w", err)
	} else if result, err := runAnalyzers(cwd, modPaths, jsPaths, s.config.Environment); err != nil {
		return fmt.Errorf("analyzers could not run: %w", err)
	} else {
		fmt.Println(result)
		return nil
	}
}

func Create(config Config) (command, error) {
	analysisCmd := flag.NewFlagSet(Name, flag.ExitOnError)

	analysisCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		analysisCmd.PrintDefaults()
	}

	if err := analysisCmd.Parse(os.Args[2:]); err != nil {
		analysisCmd.Usage()
		return command{}, fmt.Errorf("failed to parse analysis command: %w", err)
	} else {
		return command{config: config}, nil
	}
}

func runAnalyzers(cwd string, modPaths []string, jsPaths []string, env []string) (string, error) {
	golint, err := runGolangcilint(cwd, modPaths, env)
	if err != nil {
		return "", fmt.Errorf("go lint: %w", err)
	}

	eslint, err := runAllEslint(jsPaths, env)
	if err != nil {
		return "", fmt.Errorf("eslint: %w", err)
	}

	var (
		totalLen          = len(golint) + len(eslint)
		codeClimateReport = make([]codeClimateEntry, 0, totalLen)
	)

	codeClimateReport = append(codeClimateReport, golint...)
	codeClimateReport = append(codeClimateReport, eslint...)

	codeClimateReport = slices.Map(codeClimateReport, func(entry codeClimateEntry) codeClimateEntry {
		if path, err := filepath.Rel(cwd, entry.Location.Path); err != nil {
			return entry
		} else {
			entry.Location.Path = path
			return entry
		}
	})

	if jsonBytes, err := json.MarshalIndent(codeClimateReport, "", "    "); err != nil {
		return "", fmt.Errorf("could not marshal code climate report: %w", err)
	} else {
		return string(jsonBytes), nil
	}
}

func runGolangcilint(cwd string, modPaths []string, env []string) ([]codeClimateEntry, error) {
	var (
		result []codeClimateEntry
		args   = []string{"run", "--out-format", "code-climate", "--config", ".golangci.json", "--"}
		outb   bytes.Buffer
	)

	args = append(args, slices.Map(modPaths, func(modPath string) string {
		return path.Join(modPath, "...")
	})...)

	cmd := exec.Command("golangci-lint")
	cmd.Env = env
	cmd.Dir = cwd
	cmd.Stdout = &outb
	cmd.Args = append(cmd.Args, args...)

	err := cmd.Run()
	if err != nil {
		log.Printf("Ignoring non-zero exit status from golangci-lint command: %v", err)
	}

	if jsonErr := json.NewDecoder(&outb).Decode(&result); jsonErr != nil {
		return result, fmt.Errorf("failed to decode output: %w", jsonErr)
	} else {
		return result, nil
	}
}

func runAllEslint(jsPaths []string, env []string) ([]codeClimateEntry, error) {
	var result = make([]codeClimateEntry, 0, len(jsPaths))

	for _, path := range jsPaths {
		if entries, err := runEslint(path, env); err != nil {
			return result, fmt.Errorf("failed to run eslint at %v: %w", path, err)
		} else {
			result = append(result, entries...)
		}
	}

	return result, nil
}

func runEslint(cwd string, env []string) ([]codeClimateEntry, error) {
	var (
		result    []codeClimateEntry
		rawResult []esLintEntry
		outb      bytes.Buffer
	)

	cmd := exec.Command("yarn", "run", "lint", "--format", "json")
	cmd.Env = env
	cmd.Dir = cwd
	cmd.Stdout = &outb

	err := cmd.Run()
	if err != nil {
		log.Printf("Ignoring non-zero exit status from eslint command: %v", err)
	}

	err = json.NewDecoder(&outb).Decode(&rawResult)
	if err != nil {
		return result, fmt.Errorf("failed to decode output: %w", err)
	}

	for _, entry := range rawResult {
		if entry.WarningCount > 0 || entry.ErrorCount > 0 || entry.FatalErrorCount > 0 {
			for _, msg := range entry.Messages {
				var severity string

				switch msg.Severity {
				case 0:
					severity = "info"
				case 1:
					severity = "warning"
				case 2:
					severity = "critical"
				}

				ccEntry := codeClimateEntry{
					Description: msg.RuleID + ": " + msg.Message,
					Severity:    severity,
					Location: codeClimateLocation{
						Path: entry.FilePath,
						Lines: codeClimateLines{
							Begin: msg.Line,
						},
					},
				}

				result = append(result, ccEntry)
			}
		}
	}

	return result, nil
}
