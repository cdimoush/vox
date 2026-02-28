# Sub-Agents

Sub-agents are specialized agents that the main Claude agent can deploy to handle focused tasks. This keeps the main agent context clean while offloading implementation work.

## Available Agents

### spec-executor (Recommended)
**Purpose**: Beads-aware spec implementation with parallel execution support.

**When to use**:
- Implementing specs from an epic directory
- Working through beads-tracked tasks
- Multiple independent specs can run in parallel (deploy in single message)

**Key features**:
- Automatic beads task lifecycle (in_progress → closed)
- Structured report output for coordination
- Optimized for parallel Opus 4.5 instances

**See**: [spec-executor.md](./spec-executor.md)

### feature-dev (Legacy)
**Purpose**: Implement individual features or chores from spec files.

**When to use**:
- Implementing specs from `specs/` directory
- Working through an epic's individual specs
- Any focused implementation task that has a clear spec

**See**: [feature-dev.md](./feature-dev.md)

## Deployment Patterns

### Serial Deployment
Use when specs have dependencies (one must complete before the next starts).

```
Main Agent Flow:
1. Read epic README to understand execution order
2. Identify dependent specs (check "Execution Order" section)
3. Deploy feature-dev for Spec A, wait for completion
4. Review results, verify success
5. Deploy feature-dev for Spec B, wait for completion
6. Continue until all dependent specs complete
```

### Parallel Deployment
Use when specs are independent (can run simultaneously).

```
Main Agent Flow:
1. Read epic README to identify independent specs
2. Send SINGLE message with MULTIPLE Task tool calls
3. Each Task call invokes feature-dev with different spec
4. Wait for all agents to complete
5. Review all results together
```

**Critical**: For parallel execution, all Task tool calls MUST be in the same message.

## How to Invoke feature-dev

### Single Spec (Serial)
```
Use Task tool:
  subagent_type: "general-purpose"
  prompt: |
    You are implementing a spec for this project.

    Read and implement: specs/epic-name/feature-xyz.md

    Instructions:
    1. Read the spec file completely
    2. Read all files mentioned in "Relevant Files" section
    3. Follow step-by-step tasks in exact order
    4. Run validation commands at the end
    5. Report results in this format:

    ## Implementation Complete
    ### Files Created
    - path - description
    ### Files Modified
    - path - what changed
    ### Validation Results
    - command: PASS/FAIL + output
    ### Notes
    Any issues or follow-up items
```

### Multiple Specs (Parallel)
Send ONE message with multiple Task tool invocations:
```
[Task 1]
  subagent_type: "general-purpose"
  prompt: "Implement specs/epic-name/chore-a.md ..."

[Task 2]
  subagent_type: "general-purpose"
  prompt: "Implement specs/epic-name/feature-b.md ..."

[Task 3]
  subagent_type: "general-purpose"
  prompt: "Implement specs/epic-name/feature-c.md ..."
```

## Determining Serial vs Parallel

Read the epic's "Execution Order" and "Path Dependencies Diagram" sections:

**Serial indicators**:
- "Execute in order: 1, 2, 3..."
- "depends on", "requires", "must exist first"
- Arrows (↓) between specs in dependency diagram
- One spec creates files another spec imports

**Parallel indicators**:
- "Can be done in parallel"
- Specs in same phase with no dependencies
- Different platforms (macOS/Ubuntu)
- Different feature areas with no shared code

## Epic Implementation Workflow

When asked to implement an epic:

1. **Read the epic README**
   ```
   Read: specs/epic-name/README.md
   ```

2. **Parse execution order**
   - Identify phases
   - Note dependencies within each phase
   - Note user breakpoints (🔄)

3. **Implement phase by phase**
   - Deploy agents for current phase specs
   - Serial for dependent specs, parallel for independent
   - Wait for phase completion
   - Report phase results to user

4. **Respect user breakpoints**
   - Stop at 🔄 breakpoints
   - Report progress and results
   - Wait for user approval to continue

5. **Continue or complete**
   - Move to next phase after approval
   - Report final epic completion status

## Error Recovery

If an agent reports errors:

1. **Review the error** in agent's report
2. **Determine cause**:
   - Missing dependency → Install and retry
   - Unclear spec → Ask user for clarification
   - Test failure → Fix issue, retry agent
   - Blocking issue → Escalate to user

3. **Retry or skip**:
   - Retry same agent after fixing issue
   - Skip to next spec if appropriate
   - Stop if blocking for all remaining work

## Best Practices

- **Read specs before deploying** to understand dependencies
- **Deploy parallel when possible** for efficiency
- **Report progress** between phases
- **Respect breakpoints** for user validation
- **Keep main context clean** by delegating implementation details
