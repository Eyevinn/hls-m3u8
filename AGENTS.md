# AGENTS.md

Read [CONTRIBUTING.md](CONTRIBUTING.md) first. It has the code conventions, the commit message rules,
how to credit an agent as a co-author, how to add changelog entries in
[changelog.d](changelog.d/README.md) and how releases are prepared. In addition:

- End each commit message with a `Co-Authored-By:` trailer that names you and your model.
- If the pre-commit hooks are not installed, run `go fmt ./...`, `make lint`, `make test` and
  `scripts/check-changelog.sh` before committing.
- Only commit messages and pull request descriptions tell history: what was broken, why it changed,
  and measurements. Code comments, test comments and changelog entries describe the current behaviour.
