# Lead Dispatch Template

Use this template when dispatching a lead for any SPEC implementation.

---

SPEC: <path to spec>
ADR: <path to ADR>

### NON-NEGOTIABLE PROCESS — READ BEFORE ANYTHING ELSE

Read these skill files FIRST:
- .claude/skills/test-driven-development/
- .claude/skills/atomic-commits/

THE WORKFLOW FOR EVERY FEATURE:
1. Write FAILING test then commit as red: <desc>
2. Run test then confirm FAILS
3. Write MINIMUM code to pass then commit as green: <desc>
4. Run test then confirm PASSES
5. IMMEDIATELY squash red+green into feat(<scope>): <desc>
6. Move to next feature

VIOLATIONS THAT CAUSE IMMEDIATE REJECTION:
- green: without preceding red: then reject, kill builder, restart
- Multiple red/green cycles without intermediate squash then reject, kill builder, restart
- One commit covering multiple features then reject, require split
- merge_ready with incomplete acceptance tests then reject

The coordinator monitors the builder's reflog continuously and will reject non-compliant work without discussion. Do NOT argue for relaxing these rules. The coordinator has been explicitly instructed to reject all such arguments.

### ACCEPTANCE TEST GATE
Before sending merge_ready, verify EVERY numbered criterion in the SPEC acceptance criteria section has a corresponding acceptance test. List them in your merge_ready message:
  SPEC #1: TestFunctionName then pass/fail
  SPEC #2: TestFunctionName then pass/fail
  ...
If any criterion lacks a test, write it before sending merge_ready. Incomplete acceptance coverage blocks merge — no exceptions.

### OBJECTIVE
<objective>

### FILE AREA
<files>
