package audit

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

type TicketAudit struct {
	Tickets    []string
	Exceptions []string
}

type resultJSON struct {
	Body string `json:"body"`
	Url  string `json:"url"`
}

const (
	Name  = "audit"
	Usage = "Audit pull requests for ticket association"
)

var ticketRegex = regexp.MustCompile(`BED-[0-9]{4,}`)

type command struct {
	env        environment.Environment
	start      string
	end        string
	baseBranch string
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

	cmd.StringVar(&s.start, "start", "", "SHA or tag name to start from")
	cmd.StringVar(&s.end, "end", "", "SHA or tag name to end with")
	cmd.StringVar(&s.baseBranch, "base", "", "branch to use as base filter")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	if s.start == "" {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: no starting sha given", Name)
	}

	if s.end == "" {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: no ending sha given", Name)
	}

	return nil
}

// Run build command
func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if start, err := git.ParseTimestampFromSHA(s.env, paths.Root, s.start); err != nil {
		return fmt.Errorf("parsing timestamp for starting SHA: %w", err)
	} else if end, err := git.ParseTimestampFromSHA(s.env, paths.Root, s.end); err != nil {
		return fmt.Errorf("parsing timestamp for ending SHA: %w", err)
	} else if audit, err := getTicketAudit(s.env, paths.Root, start, end, s.baseBranch); err != nil {
		return fmt.Errorf("getting ticket audit: %w", err)
	} else {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		return enc.Encode(audit)
	}
}

func getTicketAudit(env environment.Environment, cwd, startTimestamp, endTimestamp, baseBranch string) (TicketAudit, error) {
	var (
		jsonBody []resultJSON

		tickets    = make([]string, 0, 64)
		exceptions = make([]string, 0, 64)
		args       = []string{"pr", "list", "--state", "merged", "--json", "url,body"}
	)

	if baseBranch != "" {
		args = append(args, "--search", fmt.Sprintf("merged:%s..%s base:%s", startTimestamp, endTimestamp, baseBranch))
	} else {
		args = append(args, "--search", fmt.Sprintf("merged:%s..%s", startTimestamp, endTimestamp))
	}

	slog.Info(
		"Ticket Audit Args",
		slog.String("startTimestamp", startTimestamp),
		slog.String("endTimestamp", endTimestamp),
		slog.String("baseBranch", baseBranch),
		slog.Any("args", args),
	)

	result, err := cmdrunner.Run("gh", args, cwd, env)
	if err != nil {
		return TicketAudit{}, fmt.Errorf("github cli PR listing: %w", err)
	}

	err = json.NewDecoder(result.StandardOutput).Decode(&jsonBody)
	if err != nil {
		return TicketAudit{}, fmt.Errorf("decoding github cli PR listing JSON: %w", err)
	}

	for _, pr := range jsonBody {
		var found bool
		matches := ticketRegex.FindAllString(pr.Body, -1)
		for _, match := range matches {
			found = true
			tickets = append(tickets, "https://specterops.atlassian.net/browse/"+match)
		}
		if !found {
			exceptions = append(exceptions, pr.Url)
		}
	}

	slices.Sort(tickets)

	return TicketAudit{
		slices.Compact(tickets),
		exceptions,
	}, nil
}
