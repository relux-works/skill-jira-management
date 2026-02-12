package main

import "github.com/ivalx1s/skill-jira-management/internal/config"

// localeStrings holds locale-specific default text templates.
var localeStrings = map[config.Locale]map[string]string{
	config.LocaleEN: {
		"dod_heading":       "Definition of Done",
		"comment_prefix":    "",
		"created_message":   "Created %s: %s/browse/%s",
		"transition_done":   "%s -> %s",
		"comment_added":     "Comment added to %s (id: %s)",
		"dod_set":           "DoD set on %s (comment id: %s)",
	},
	config.LocaleRU: {
		"dod_heading":       "Критерии приёмки",
		"comment_prefix":    "",
		"created_message":   "Создано %s: %s/browse/%s",
		"transition_done":   "%s -> %s",
		"comment_added":     "Комментарий добавлен к %s (id: %s)",
		"dod_set":           "DoD установлен на %s (comment id: %s)",
	},
}

// getLocaleString returns a locale-aware string, falling back to English.
func getLocaleString(locale config.Locale, key string) string {
	if strs, ok := localeStrings[locale]; ok {
		if s, ok := strs[key]; ok {
			return s
		}
	}
	// Fallback to EN
	if strs, ok := localeStrings[config.LocaleEN]; ok {
		if s, ok := strs[key]; ok {
			return s
		}
	}
	return key
}

// getConfigLocale reads the locale from config, defaults to EN.
func getConfigLocale() config.Locale {
	cfgMgr, err := config.NewConfigManager()
	if err != nil {
		return config.LocaleEN
	}
	cfg, err := cfgMgr.GetConfig()
	if err != nil {
		return config.LocaleEN
	}
	if cfg.Locale == "" {
		return config.LocaleEN
	}
	return cfg.Locale
}
