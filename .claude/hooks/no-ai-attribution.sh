#!/bin/bash
#
# Pre-tool-use hook that blocks any Bash command containing AI agent
# attribution (e.g. "Co-Authored-By: Claude", "Generated with Claude Code").
#
# This enforces the AGENTS.md rule:
#   "Do NOT add the agent name anywhere in commit messages or commit usernames."

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -z "$COMMAND" ]; then
  exit 0
fi

if echo "$COMMAND" | grep -qi 'co-authored-by.*claude\|generated with.*claude\|claude code\|noreply@anthropic'; then
  jq -n '{
    "hookSpecificOutput": {
      "hookEventName": "PreToolUse",
      "permissionDecision": "deny",
      "permissionDecisionReason": "BLOCKED: Your command contains AI agent attribution. Per AGENTS.md, do NOT add the agent name (e.g. Claude, Generated with Claude Code, Co-Authored-By Claude) anywhere in commit messages, PR descriptions, or other Git/GitHub messages. Remove the attribution and try again."
    }
  }'
  exit 0
fi

exit 0
