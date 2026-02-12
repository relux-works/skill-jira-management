# Multi-Step Workflow Patterns

Complex workflows for `jira-mgmt` CLI. Step-by-step patterns for common agent tasks.

---

## Epic Creation with Stories

### Pattern: Create Epic → Create Stories → Link

**Scenario:** Create epic "User Authentication" with 3 stories.

**Steps:**

```bash
# Step 1: Create epic
EPIC=$(jira-mgmt create \
  --type epic \
  --summary "User Authentication System" \
  --description "Implement OAuth2 and local authentication" \
  --project PROJ \
  --format json | jq -r '.key')

echo "Created epic: $EPIC"

# Step 2: Create stories under epic
STORY1=$(jira-mgmt create \
  --type story \
  --summary "Google OAuth2 integration" \
  --parent "$EPIC" \
  --project PROJ \
  --format json | jq -r '.key')

STORY2=$(jira-mgmt create \
  --type story \
  --summary "GitHub OAuth2 integration" \
  --parent "$EPIC" \
  --project PROJ \
  --format json | jq -r '.key')

STORY3=$(jira-mgmt create \
  --type story \
  --summary "Local authentication fallback" \
  --parent "$EPIC" \
  --project PROJ \
  --format json | jq -r '.key')

echo "Created stories: $STORY1, $STORY2, $STORY3"

# Step 3: Verify epic hierarchy
jira-mgmt q "get($EPIC){full}"
```

**Expected Output:**
```
Created epic: PROJ-100
Created stories: PROJ-101, PROJ-102, PROJ-103

PROJ-100: User Authentication System
Type: Epic
Status: To Do

Description:
Implement OAuth2 and local authentication

Children (3):
- PROJ-101: Google OAuth2 integration
- PROJ-102: GitHub OAuth2 integration
- PROJ-103: Local authentication fallback
```

---

## Bulk Status Transitions

### Pattern: Query Issues → Transition All

**Scenario:** Start all "To Do" issues in current sprint.

**Steps:**

```bash
# Step 1: Get all "To Do" issues in current sprint
ISSUES=$(jira-mgmt q 'list(sprint=current,status=todo){minimal}' --format json | jq -r '.[].key')

# Step 2: Count issues
COUNT=$(echo "$ISSUES" | wc -l)
echo "Found $COUNT issues to transition"

# Step 3: Transition each to "In Progress"
for issue in $ISSUES; do
  echo "Transitioning $issue..."
  jira-mgmt transition "$issue" --to "In Progress"
  jira-mgmt comment "$issue" --body "Batch start: Sprint 24"
done

# Step 4: Verify
jira-mgmt q 'list(sprint=current,status=in-progress){default}'
```

**Expected Output:**
```
Found 5 issues to transition
Transitioning PROJ-123...
Transitioning PROJ-124...
Transitioning PROJ-125...
Transitioning PROJ-126...
Transitioning PROJ-127...

Sprint: Sprint 24 (5 in progress)

PROJ-123: Story: User Authentication [In Progress] @alice
PROJ-124: Story: Password Reset [In Progress] @bob
PROJ-125: Task: Write tests [In Progress] @charlie
PROJ-126: Bug: Login fails [In Progress] @alice
PROJ-127: Task: Add validation [In Progress] @bob
```

---

## Story Decomposition

### Pattern: Create Story → Analyze → Create Tasks → Transition to Dev

**Scenario:** Create story with sub-tasks based on analysis.

**Steps:**

```bash
# Step 1: Create story
STORY=$(jira-mgmt create \
  --type story \
  --summary "Payment gateway integration" \
  --description "Integrate Stripe and PayPal payment gateways" \
  --project PROJ \
  --format json | jq -r '.key')

echo "Created story: $STORY"

# Step 2: Add analysis comment
jira-mgmt comment "$STORY" --body "$(cat <<EOF
Analysis:
- Need Stripe SDK setup
- Need PayPal SDK setup
- Implement payment flow for both
- Add webhook handlers
- Write integration tests
EOF
)"

# Step 3: Create sub-tasks based on analysis
TASK1=$(jira-mgmt create \
  --type subtask \
  --summary "Add Stripe SDK" \
  --parent "$STORY" \
  --project PROJ \
  --format json | jq -r '.key')

TASK2=$(jira-mgmt create \
  --type subtask \
  --summary "Add PayPal SDK" \
  --parent "$STORY" \
  --project PROJ \
  --format json | jq -r '.key')

TASK3=$(jira-mgmt create \
  --type subtask \
  --summary "Implement Stripe payment flow" \
  --parent "$STORY" \
  --project PROJ \
  --format json | jq -r '.key')

TASK4=$(jira-mgmt create \
  --type subtask \
  --summary "Implement PayPal payment flow" \
  --parent "$STORY" \
  --project PROJ \
  --format json | jq -r '.key')

TASK5=$(jira-mgmt create \
  --type subtask \
  --summary "Add webhook handlers" \
  --parent "$STORY" \
  --project PROJ \
  --format json | jq -r '.key')

TASK6=$(jira-mgmt create \
  --type subtask \
  --summary "Write integration tests" \
  --parent "$STORY" \
  --project PROJ \
  --format json | jq -r '.key')

# Step 4: Set DoD for story
jira-mgmt dod "$STORY" --set "$(cat <<EOF
All sub-tasks completed
Stripe and PayPal integration working
Webhook handlers tested
Integration tests pass
Documentation updated
EOF
)"

# Step 5: Transition story to "Ready for Dev"
jira-mgmt transition "$STORY" --to "Ready for Dev"

# Step 6: Verify story with sub-tasks
jira-mgmt q "get($STORY){full}"
```

---

## Research to Task Flow

### Pattern: Create Research Task → Add Findings → Create Implementation Tasks

**Scenario:** Research OAuth2 providers, then create implementation tasks.

**Steps:**

```bash
# Step 1: Create research task
RESEARCH=$(jira-mgmt create \
  --type task \
  --summary "Research OAuth2 providers" \
  --description "Evaluate Google, GitHub, and Auth0 for OAuth2 integration" \
  --project PROJ \
  --format json | jq -r '.key')

# Step 2: Transition to "In Progress"
jira-mgmt transition "$RESEARCH" --to "In Progress"

# Step 3: Add research findings as comment
jira-mgmt comment "$RESEARCH" --body "$(cat <<EOF
Research Findings:

Google OAuth2:
- SDK: @google-auth/library
- Pros: Well documented, widely used
- Cons: Requires Google Cloud project setup
- Effort: 3 days

GitHub OAuth2:
- SDK: Built-in REST API
- Pros: Simple, no external SDK
- Cons: Limited scope control
- Effort: 2 days

Auth0:
- SDK: auth0-js
- Pros: Multi-provider support, enterprise features
- Cons: Paid service, vendor lock-in
- Effort: 5 days

Recommendation: Google + GitHub (5 days total)
EOF
)"

# Step 4: Mark research complete
jira-mgmt transition "$RESEARCH" --to "Done"

# Step 5: Create implementation tasks based on findings
IMPL1=$(jira-mgmt create \
  --type task \
  --summary "Implement Google OAuth2" \
  --description "3 day effort. Use @google-auth/library" \
  --project PROJ \
  --format json | jq -r '.key')

IMPL2=$(jira-mgmt create \
  --type task \
  --summary "Implement GitHub OAuth2" \
  --description "2 day effort. Use GitHub REST API" \
  --project PROJ \
  --format json | jq -r '.key')

# Step 6: Link research to implementation tasks
jira-mgmt comment "$IMPL1" --body "Based on research: $RESEARCH"
jira-mgmt comment "$IMPL2" --body "Based on research: $RESEARCH"

# Step 7: Verify
jira-mgmt q "get($RESEARCH){overview}; get($IMPL1){overview}; get($IMPL2){overview}"
```

---

## Sprint Planning Workflow

### Pattern: Review Backlog → Select Stories → Move to Sprint

**Scenario:** Plan Sprint 25 with 10 story points.

**Steps:**

```bash
# Step 1: Get backlog items (not in sprint, ready for dev)
echo "=== Backlog Items ==="
jira-mgmt q 'search(jql="sprint is EMPTY AND status=Ready for Dev"){overview}'

# Step 2: Select stories for sprint (manual selection, then script)
STORIES=(
  "PROJ-150"  # 3 points
  "PROJ-151"  # 5 points
  "PROJ-152"  # 2 points
)

# Step 3: Add comment to each (planning notes)
for story in "${STORIES[@]}"; do
  jira-mgmt comment "$story" --body "Added to Sprint 25 planning"
done

# Step 4: Verify sprint scope
echo "=== Sprint 25 Scope ==="
for story in "${STORIES[@]}"; do
  jira-mgmt q "get($story){overview}"
done

# Step 5: Get current sprint summary
jira-mgmt q 'summary()'
```

**Note:** Actual sprint assignment requires Jira API sprint operations (not yet implemented in CLI).

---

## Sprint Review Workflow

### Pattern: Summary → Completed Work → Carry-Over → Report

**Scenario:** Generate sprint review report.

**Steps:**

```bash
# Step 1: Create report file
REPORT="/tmp/sprint-24-review.txt"

# Step 2: Add sprint summary
echo "=== Sprint 24 Review ===" > "$REPORT"
echo "" >> "$REPORT"
jira-mgmt q 'summary()' >> "$REPORT"

# Step 3: Add completed work
echo "" >> "$REPORT"
echo "=== Completed Work ===" >> "$REPORT"
jira-mgmt q 'list(sprint=current,status=done){overview}' >> "$REPORT"

# Step 4: Add carry-over items
echo "" >> "$REPORT"
echo "=== Carry-Over Items ===" >> "$REPORT"
jira-mgmt q 'list(sprint=current,status=!done){default}' >> "$REPORT"

# Step 5: Add velocity metrics
echo "" >> "$REPORT"
echo "=== Velocity ===" >> "$REPORT"
COMPLETED=$(jira-mgmt q 'list(sprint=current,status=done){minimal}' --format json | jq '. | length')
TOTAL=$(jira-mgmt q 'list(sprint=current){minimal}' --format json | jq '. | length')
echo "Completed: $COMPLETED / $TOTAL issues" >> "$REPORT"

# Step 6: View report
cat "$REPORT"

# Step 7: Add report as comment to sprint epic (if exists)
SPRINT_EPIC="PROJ-100"  # Replace with actual sprint epic
jira-mgmt comment "$SPRINT_EPIC" --body "$(cat $REPORT)"
```

---

## Epic Decomposition Workflow

### Pattern: Create Epic → Create Stories with Sub-tasks

**Scenario:** Create "Payment System" epic with full hierarchy.

**Steps:**

```bash
# Step 1: Create parent epic
EPIC=$(jira-mgmt create \
  --type epic \
  --summary "Payment System" \
  --description "Complete payment processing system with multiple providers" \
  --project PROJ \
  --format json | jq -r '.key')

echo "Created epic: $EPIC"

# Step 2: Create stories under epic
STORY1=$(jira-mgmt create \
  --type story \
  --summary "Stripe integration" \
  --parent "$EPIC" \
  --project PROJ \
  --format json | jq -r '.key')

STORY2=$(jira-mgmt create \
  --type story \
  --summary "PayPal integration" \
  --parent "$EPIC" \
  --project PROJ \
  --format json | jq -r '.key')

STORY3=$(jira-mgmt create \
  --type story \
  --summary "Payment history UI" \
  --parent "$EPIC" \
  --project PROJ \
  --format json | jq -r '.key')

# Step 3: Create sub-tasks for Story 1 (Stripe)
jira-mgmt create --type subtask --summary "Add Stripe SDK" --parent "$STORY1" --project PROJ
jira-mgmt create --type subtask --summary "Implement payment flow" --parent "$STORY1" --project PROJ
jira-mgmt create --type subtask --summary "Add webhook handlers" --parent "$STORY1" --project PROJ

# Step 4: Create sub-tasks for Story 2 (PayPal)
jira-mgmt create --type subtask --summary "Add PayPal SDK" --parent "$STORY2" --project PROJ
jira-mgmt create --type subtask --summary "Implement payment flow" --parent "$STORY2" --project PROJ

# Step 5: Create sub-tasks for Story 3 (UI)
jira-mgmt create --type subtask --summary "Design payment history page" --parent "$STORY3" --project PROJ
jira-mgmt create --type subtask --summary "Implement payment list component" --parent "$STORY3" --project PROJ
jira-mgmt create --type subtask --summary "Add filters and search" --parent "$STORY3" --project PROJ

# Step 6: View epic hierarchy
jira-mgmt q "get($EPIC){full}"
```

---

## Issue Progression with DoD

### Pattern: Create → Set DoD → Progress → Verify → Complete

**Scenario:** Full lifecycle of a task with Definition of Done.

**Steps:**

```bash
# Step 1: Create task
TASK=$(jira-mgmt create \
  --type task \
  --summary "Implement OAuth2 provider" \
  --description "Add OAuth2 authentication with Google" \
  --project PROJ \
  --assignee alice@example.com \
  --format json | jq -r '.key')

# Step 2: Set Definition of Done
jira-mgmt dod "$TASK" --set "$(cat <<EOF
Unit tests pass
Integration tests pass
Code reviewed by 2 team members
Security audit complete
Documentation updated
Deployed to staging
EOF
)"

# Step 3: Transition to "In Progress"
jira-mgmt transition "$TASK" --to "In Progress"

# Step 4: Add progress updates
jira-mgmt comment "$TASK" --body "Started implementation. Added OAuth2 SDK."
sleep 2
jira-mgmt comment "$TASK" --body "Implemented authentication flow. Tests pending."
sleep 2
jira-mgmt comment "$TASK" --body "Unit tests complete. Starting integration tests."

# Step 5: Transition to "In Review"
jira-mgmt transition "$TASK" --to "In Review"
jira-mgmt comment "$TASK" --body "@bob@example.com @charlie@example.com please review"

# Step 6: Transition to "Done"
jira-mgmt transition "$TASK" --to "Done"
jira-mgmt comment "$TASK" --body "$(cat <<EOF
All DoD criteria met:
✓ Unit tests pass
✓ Integration tests pass
✓ Code reviewed by Bob and Charlie
✓ Security audit complete
✓ Documentation updated
✓ Deployed to staging

Ready for production deployment.
EOF
)"

# Step 7: Verify final state
jira-mgmt q "get($TASK){full}"
```

---

## Daily Standup Report

### Pattern: Query Yesterday → Query Today → Query Blockers → Report

**Scenario:** Generate daily standup report.

**Steps:**

```bash
# Step 1: Create report file
REPORT="/tmp/standup-$(date +%Y-%m-%d).txt"

# Step 2: Add header
echo "=== Daily Standup - $(date +%Y-%m-%d) ===" > "$REPORT"
echo "" >> "$REPORT"

# Step 3: Add yesterday's completed work
echo "## Completed Yesterday" >> "$REPORT"
jira-mgmt q 'search(jql="assignee=currentUser() AND statusCategoryChangedDate>=-1d AND statusCategory=Done"){default}' >> "$REPORT"

# Step 4: Add today's in progress
echo "" >> "$REPORT"
echo "## In Progress Today" >> "$REPORT"
jira-mgmt q 'list(assignee=me,status=in-progress){default}' >> "$REPORT"

# Step 5: Add blockers
echo "" >> "$REPORT"
echo "## Blocked Issues" >> "$REPORT"
jira-mgmt q 'search(jql="assignee=currentUser() AND status=Blocked"){default}' >> "$REPORT"

# Step 6: Display report
cat "$REPORT"
```

**Expected Output:**
```
=== Daily Standup - 2026-02-12 ===

## Completed Yesterday
PROJ-120: Task: Add OAuth2 SDK [Done] @alice
PROJ-121: Bug: Fix login redirect [Done] @alice

## In Progress Today
PROJ-123: Story: User Authentication [In Progress] @alice
PROJ-125: Task: Write integration tests [In Progress] @alice

## Blocked Issues
(none)
```

---

## Bug Triage Workflow

### Pattern: Query Bugs → Prioritize → Assign → Track

**Scenario:** Triage new bugs from last 7 days.

**Steps:**

```bash
# Step 1: Get new bugs
echo "=== New Bugs (Last 7 Days) ==="
BUGS=$(jira-mgmt q 'search(jql="issuetype=Bug AND created>=-7d"){default}' --format json)

# Step 2: Count bugs
COUNT=$(echo "$BUGS" | jq '. | length')
echo "Found $COUNT new bugs"

# Step 3: Display bugs
echo "$BUGS" | jq -r '.[] | "\(.key): \(.summary) [\(.priority)]"'

# Step 4: Assign high priority bugs (example)
HIGH_PRIORITY_BUGS=$(echo "$BUGS" | jq -r '.[] | select(.priority == "High" or .priority == "Highest") | .key')

for bug in $HIGH_PRIORITY_BUGS; do
  echo "Assigning high priority bug: $bug"
  jira-mgmt comment "$bug" --body "High priority bug - assigning to team lead for triage"
done

# Step 5: Create triage report
echo "=== Bug Triage Report ===" > /tmp/bug-triage.txt
jira-mgmt q 'search(jql="issuetype=Bug AND created>=-7d"){overview}' >> /tmp/bug-triage.txt
cat /tmp/bug-triage.txt
```

---

## Batch Operations

### Pattern: Extract Keys → Iterate → Apply Operation

**Generic pattern for bulk operations:**

```bash
# Extract issue keys from query
KEYS=$(jira-mgmt q 'list(sprint=current,status=todo)' --format json | jq -r '.[].key')

# Iterate and apply operation
for key in $KEYS; do
  # Example: add label
  jira-mgmt comment "$key" --body "Adding sprint-24 label"
  # Note: label update not yet in CLI, use comment as example
done
```

---

## Critical Path Analysis

### Pattern: Fetch Stories with Subtasks → Map Dependencies → Build Critical Path → Phase Plan

**Scenario:** User has multiple stories assigned. Analyze all subtasks, find cross-story dependencies, build the critical path and recommend execution phases.

**Steps:**

```bash
# Step 1: Fetch all assigned stories with subtasks
jira-mgmt q 'search(jql="assignee=currentUser() AND issuetype=Story AND NOT statusCategory = Done"){full}'
```

**Step 2: Group subtasks by story**

Present a table for each story:

```
## STORY-1: Story title

| # | Key | Component | Task | Status |
|---|-----|-----------|------|--------|
| 1 | PROJ-10 | Backend | Implement API | Open |
| 2 | PROJ-11 | Frontend | Build UI | Open |
| ... |
```

**Step 3: Map dependencies as ASCII graph**

Analyze subtasks by component tags (e.g. `[Arch]`, `[BE]`, `[SDK]`, `[Infra]`) and domain logic to infer execution order. Draw an ASCII dependency graph showing both stories:

```
STORY-1 (Feature A)                    STORY-2 (Feature B)
========================               ========================

PROJ-10 [Arch]
  Architecture
       │
       ├──────────────┐
       ▼              ▼
PROJ-11 [Infra]    PROJ-12 [BE]
  Infrastructure     Backend core
       │              │
       │              ▼
       │         PROJ-13 [BE] ──────► PROJ-20 [BE]
       │           Service A            Service B (Story 2)
       │                                    │
       ▼                                    ▼
PROJ-14 [HLR]                         PROJ-21 [SDK]
  Routing                                Client SDK
```

Key rules:
- Show cross-story dependencies with `──────►` arrows
- Mark parallel tracks (tasks that can run simultaneously)
- Identify shared components between stories

**Step 4: Identify the critical path**

The critical path is the **longest sequential chain** through the dependency graph. List it explicitly:

```
### Critical Path (longest chain)

PROJ-10 → PROJ-12 → PROJ-13 → PROJ-20 → PROJ-21 → PROJ-22
Arch      BE core    Service    Service    SDK        Integration
                     A          B          client     test

6 tasks in sequence. This determines the minimum project duration.
```

**Step 5: Identify parallel tracks**

Tasks NOT on the critical path that can run alongside it:

```
### Parallel Tracks (not on critical path)

| Task | Can start after | Runs parallel to |
|------|----------------|------------------|
| PROJ-11 [Infra] | PROJ-10 (Arch) | PROJ-12 and everything after |
| PROJ-14 [HLR] | PROJ-10 (Arch) | Entire BE track |
```

**Step 6: Recommend execution phases**

Group tasks into phases based on dependencies. Each phase can start only when its prerequisites are done:

```
### Execution Phases

**Phase 1 — Foundation (can parallelize):**
- PROJ-10 [Arch] — blocks everything, start first
- PROJ-11 [Infra] — start immediately after/with arch

**Phase 2 — Core backend (sequential):**
- PROJ-12 → PROJ-13 → PROJ-14

**Phase 3 — Feature backend + routing (parallel):**
- PROJ-20 → PROJ-21 (Story 2 backend)
- PROJ-15 [HLR] (independent)

**Phase 4 — Client-side (sequential):**
- PROJ-22 → PROJ-23 → PROJ-24
```

### Tips for Critical Path Analysis

- **Component tags in subtask titles** (`[Arch]`, `[BE]`, `[SDK]`, `[Infra]`, `[HLR]`) reveal the layer and help infer dependencies
- **Cross-story shared components** (e.g. SPKI registry used by both sync and auth) create the longest chains — always look for these
- **The critical path determines minimum duration** — shortening it requires parallelizing or removing tasks from the chain
- **Parallel tracks are free optimization** — identify them to keep multiple workstreams busy
- **Phase boundaries = dependency gates** — a phase starts only when all its prerequisites from the previous phase are done

---

**Document Version:** 1.1
**Last Updated:** 2026-02-12
