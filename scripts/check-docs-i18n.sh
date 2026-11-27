#!/usr/bin/env bash
# Every root-level English doc must have a Russian pair (*.ru.md) and vice
# versa. Docs inside component folders are not checked.
set -euo pipefail
files="$(git ls-files -- ':(glob)*.md')" # fails loudly outside a work tree
status=0
while IFS= read -r f; do
  [[ -n "$f" ]] || continue
  case "$f" in
    LICENSE.md) continue ;;
    *.ru.md) pair="${f%.ru.md}.md" ;;
    *) pair="${f%.md}.ru.md" ;;
  esac
  if [[ ! -f "$pair" ]]; then
    echo "missing translation pair: $f -> $pair"
    status=1
  fi
done <<< "$files"
exit $status
