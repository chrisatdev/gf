package commit

import (
	"fmt"
	"strings"
)

type Generator struct {
	detector *Detector
}

func NewGenerator() *Generator {
	return &Generator{detector: NewDetector()}
}

type CommitInfo struct {
	Type        string
	Emoji       string
	Description string
	FileCount   int
	NewFiles    int
	ModFiles    int
	DelFiles    int
}

func (g *Generator) Generate(files []string, diffContent string) *CommitInfo {
	commitType := g.detector.Detect(files, diffContent)
	emoji := GetEmoji(commitType)

	fileCount := len(files)
	newFiles := countByType(files, "new")
	modFiles := countByType(files, "modified")
	delFiles := countByType(files, "deleted")

	description := generateDescription(commitType, fileCount)

	return &CommitInfo{
		Type:        commitType,
		Emoji:       emoji,
		Description: description,
		FileCount:   fileCount,
		NewFiles:    newFiles,
		ModFiles:    modFiles,
		DelFiles:    delFiles,
	}
}

func (g *Generator) FormatMessage(info *CommitInfo) string {
	header := fmt.Sprintf("%s %s: %s", info.Emoji, info.Type, info.Description)

	var details []string
	if info.NewFiles > 0 {
		details = append(details, fmt.Sprintf("%d new", info.NewFiles))
	}
	if info.ModFiles > 0 {
		details = append(details, fmt.Sprintf("%d modified", info.ModFiles))
	}
	if info.DelFiles > 0 {
		details = append(details, fmt.Sprintf("%d deleted", info.DelFiles))
	}

	if len(details) > 0 {
		header += fmt.Sprintf("\n\nChanges: %s", strings.Join(details, ", "))
	}

	return header
}

func generateDescription(commitType string, count int) string {
	filesStr := "file"
	if count > 1 {
		filesStr = "files"
	}

	switch commitType {
	case "feat":
		return fmt.Sprintf("add new features (%d %s)", count, filesStr)
	case "fix":
		return fmt.Sprintf("resolve issues (%d %s)", count, filesStr)
	case "docs":
		return fmt.Sprintf("update documentation (%d %s)", count, filesStr)
	case "style":
		return fmt.Sprintf("improve styling (%d %s)", count, filesStr)
	case "test":
		return fmt.Sprintf("update tests (%d %s)", count, filesStr)
	case "build":
		return fmt.Sprintf("update build config (%d %s)", count, filesStr)
	case "refactor":
		return fmt.Sprintf("refactor code (%d %s)", count, filesStr)
	default:
		return fmt.Sprintf("update codebase (%d %s)", count, filesStr)
	}
}

func countByType(files []string, typeName string) int {
	// Simplified - just return len(files) as "modified"
	return len(files)
}
