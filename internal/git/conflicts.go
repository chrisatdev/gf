package git

import (
	"fmt"
	"strings"
)

type Conflict struct {
	File   string
	Ours   string
	Theirs string
}

func (r *Runner) HasConflicts() bool {
	output, _ := r.Command("diff", "--name-only", "--diff-filter=U")
	return len(output) > 0
}

func (r *Runner) GetConflictingFiles() ([]string, error) {
	output, err := r.Command("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	files := strings.Split(output, "\n")
	result := make([]string, 0, len(files))
	for _, f := range files {
		if f = strings.TrimSpace(f); f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}

func ParseConflictFile(content string) (*Conflict, error) {
	// Simple conflict marker parsing
	lines := strings.Split(content, "\n")

	var ours, theirs []string
	state := "before" // before, ours, theirs

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "<<<<<<<"):
			state = "ours"
		case strings.HasPrefix(line, "======="):
			state = "theirs"
		case strings.HasPrefix(line, ">>>>>>>"):
			state = "done"
		default:
			switch state {
			case "ours":
				ours = append(ours, line)
			case "theirs":
				theirs = append(theirs, line)
			}
		}
	}

	if len(ours) == 0 && len(theirs) == 0 {
		return nil, fmt.Errorf("no conflict markers found")
	}

	return &Conflict{
		Ours:   strings.Join(ours, "\n"),
		Theirs: strings.Join(theirs, "\n"),
	}, nil
}

func AcceptOurs(file string) error {
	r := NewRunner()
	_, err := r.Command("checkout", "--ours", file)
	return err
}

func AcceptTheirs(file string) error {
	r := NewRunner()
	_, err := r.Command("checkout", "--theirs", file)
	return err
}

func StageFile(file string) error {
	r := NewRunner()
	return r.Add(file)
}
