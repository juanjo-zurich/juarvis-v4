#!/usr/bin/env python3
"""Hookify PostToolUse hook wrapper.

Reads hook input from stdin, evaluates rules, and returns exit codes:
  0 = allow (with optional warning on stderr)
  2 = deny (blocked)
"""

import json
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from hookify.config_loader import load_rules
from hookify.rule_engine import RuleEngine


def main():
    try:
        input_data = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON from stdin: {e}", file=sys.stderr)
        sys.exit(1)

    input_data["hook_event_name"] = "PostToolUse"

    rules = load_rules(event="bash") + load_rules(event="file") + load_rules(event="all")

    if not rules:
        sys.exit(0)

    engine = RuleEngine()
    result = engine.evaluate_rules(rules, input_data)

    if not result:
        sys.exit(0)

    if "systemMessage" in result:
        print(result["systemMessage"], file=sys.stderr)

    hook_output = result.get("hookSpecificOutput", {})
    if hook_output.get("permissionDecision") == "deny":
        sys.exit(2)

    sys.exit(0)


if __name__ == "__main__":
    main()
