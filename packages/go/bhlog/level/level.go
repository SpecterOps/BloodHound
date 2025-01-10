package level

import "log/slog"

var lvl = new(slog.LevelVar)

func GetLevelVar() *slog.LevelVar {
	return lvl
}

func SetGlobalLevel(level slog.Level) {
	lvl.Set(level)
}

func GlobalAccepts(level slog.Level) bool {
	return lvl.Level() <= level
}

func GlobalLevel() slog.Level {
	return lvl.Level()
}
