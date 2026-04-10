package commit

var Gitmoji = map[string]string{
	"feat":     "✨",
	"fix":      "🐛",
	"docs":     "📝",
	"style":    "💄",
	"refactor": "♻️",
	"test":     "✅",
	"chore":    "🔧",
	"build":    "👷",
	"ci":       "⚙️",
	"perf":     "⚡",
	"revert":   "⏪",
}

func GetEmoji(commitType string) string {
	if emoji, ok := Gitmoji[commitType]; ok {
		return emoji
	}
	return "🔧"
}
