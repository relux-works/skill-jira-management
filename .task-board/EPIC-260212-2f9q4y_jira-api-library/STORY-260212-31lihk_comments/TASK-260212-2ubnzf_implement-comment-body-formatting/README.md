# TASK-260212-2ubnzf: Implement comment body formatting

## Description
Implement ADF (Atlassian Document Format) helpers: convert plain text to ADF, parse ADF to plain text. Handle basic formatting (paragraphs, bold, italic, links)

## Scope
(define task scope)

## Acceptance Criteria
- TextToADF(text) function converts plain text to ADF structure\n- ADFToText(adf) function extracts plain text from ADF\n- Handles paragraphs, line breaks\n- Handles basic formatting: bold, italic, links\n- ADF struct defined with proper nesting (doc, paragraph, text nodes)
