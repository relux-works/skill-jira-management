# degitize-installed-skill-artifacts-and-prune-source-files

## Description
Make setup install a degitized artifact copy of the skill into global environments

## Scope
Strip .git and source-only directories/files from installed skill copies and avoid layouts that make agents think globals are the source of truth

## Acceptance Criteria
Global skill installs look like runtime artifacts rather than repos; agents do not infer that edits should happen in globals; setup enforces the cleanup automatically
