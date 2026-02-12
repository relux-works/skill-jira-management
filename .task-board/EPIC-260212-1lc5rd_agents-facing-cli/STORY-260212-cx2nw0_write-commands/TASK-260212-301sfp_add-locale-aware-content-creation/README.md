# TASK-260212-301sfp: Add locale-aware content creation

## Description
Use config locale for default text in issue creation. Generate summaries/descriptions in configured locale. Support at least en_US and ru_RU. Locale setting from config.

## Scope
(define task scope)

## Acceptance Criteria
- Reads locale from config (e.g., en_US, ru_RU)
- Generates default text in configured locale
- Supports at least English and Russian
- Falls back to English if locale not supported
- Locale applies to summaries and descriptions
