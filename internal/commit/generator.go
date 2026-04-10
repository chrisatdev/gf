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
	Type            string
	Emoji           string
	Description     string
	FileCount       int
	NewFiles        int
	ModFiles        int
	DelFiles        int
	RenamedFiles    int
	NewFileList     []string
	ModFileList     []string
	DelFileList     []string
	RenamedFileList []string
}

func (g *Generator) Generate(files []string, diffContent string) *CommitInfo {
	commitType := g.detector.Detect(files, diffContent)
	emoji := GetEmoji(commitType)

	fileCount := len(files)

	description := generateDescription(commitType, fileCount)

	return &CommitInfo{
		Type:        commitType,
		Emoji:       emoji,
		Description: description,
		FileCount:   fileCount,
		NewFileList: files, // Default all as modified for now
		ModFileList: files,
	}
}

// GenerateFromFileStatus creates CommitInfo from git status
func (g *Generator) GenerateFromFileStatus(
	newFiles, modFiles, delFiles, renamedFiles []string,
	diffContent string,
) *CommitInfo {
	// Combine all files for type detection
	var allFiles []string
	allFiles = append(allFiles, newFiles...)
	allFiles = append(allFiles, modFiles...)
	allFiles = append(allFiles, delFiles...)
	allFiles = append(allFiles, renamedFiles...)

	commitType := g.detector.Detect(allFiles, diffContent)
	emoji := GetEmoji(commitType)

	fileCount := len(allFiles)
	description := generateDescription(commitType, fileCount)

	return &CommitInfo{
		Type:            commitType,
		Emoji:           emoji,
		Description:     description,
		FileCount:       fileCount,
		NewFiles:        len(newFiles),
		ModFiles:        len(modFiles),
		DelFiles:        len(delFiles),
		RenamedFiles:    len(renamedFiles),
		NewFileList:     newFiles,
		ModFileList:     modFiles,
		DelFileList:     delFiles,
		RenamedFileList: renamedFiles,
	}
}

func (g *Generator) FormatMessage(info *CommitInfo) string {
	var lines []string

	// Header: ✨ feat: add new features (3 files)
	header := fmt.Sprintf("%s %s: %s", info.Emoji, info.Type, info.Description)
	lines = append(lines, header)

	// Changes summary: Changes: 2 new, 1 modified
	var changes []string
	if info.NewFiles > 0 {
		changes = append(changes, fmt.Sprintf("%d new", info.NewFiles))
	}
	if info.ModFiles > 0 {
		changes = append(changes, fmt.Sprintf("%d modified", info.ModFiles))
	}
	if info.DelFiles > 0 {
		changes = append(changes, fmt.Sprintf("%d deleted", info.DelFiles))
	}
	if info.RenamedFiles > 0 {
		changes = append(changes, fmt.Sprintf("%d renamed", info.RenamedFiles))
	}

	if len(changes) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Changes: "+strings.Join(changes, ", "))
	}

	// New files list
	if len(info.NewFileList) > 0 {
		lines = append(lines, "")
		lines = append(lines, "**New files:**")
		for _, f := range info.NewFileList {
			lines = append(lines, fmt.Sprintf("- %s", f))
		}
	}

	// Modified files list
	if len(info.ModFileList) > 0 {
		lines = append(lines, "")
		lines = append(lines, "**Modified files:**")
		for _, f := range info.ModFileList {
			lines = append(lines, fmt.Sprintf("- %s", f))
		}
	}

	// Deleted files list
	if len(info.DelFileList) > 0 {
		lines = append(lines, "")
		lines = append(lines, "**Deleted files:**")
		for _, f := range info.DelFileList {
			lines = append(lines, fmt.Sprintf("- %s", f))
		}
	}

	// Renamed files list
	if len(info.RenamedFileList) > 0 {
		lines = append(lines, "")
		lines = append(lines, "**Renamed files:**")
		for _, f := range info.RenamedFileList {
			lines = append(lines, fmt.Sprintf("- %s", f))
		}
	}

	return strings.Join(lines, "\n")
}

func (g *Generator) FormatShortMessage(info *CommitInfo) string {
	return fmt.Sprintf("%s %s: %s", info.Emoji, info.Type, info.Description)
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
	case "ci":
		return fmt.Sprintf("update CI/CD (%d %s)", count, filesStr)
	case "perf":
		return fmt.Sprintf("improve performance (%d %s)", count, filesStr)
	default:
		return fmt.Sprintf("update codebase (%d %s)", count, filesStr)
	}
}
