package app

import "strings"

// Explain returns an actionable hint for common Git failures, or an empty
// string when the error is not recognized.
//
// Repora never runs these commands for the user: the hint tells them what to do
// in a terminal, because resolving an in-progress operation is a decision.
func Explain(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "cannot switch branch while cherry-picking"):
		return "A cherry-pick sequence is pending. If it is stale, stop it with `git cherry-pick --quit`; to finish it, run `git cherry-pick --continue`."
	case strings.Contains(msg, "cannot switch branch while merging"):
		return "A merge is in progress. Resolve it and run `git merge --continue`, or cancel with `git merge --abort`."
	case strings.Contains(msg, "cannot switch branch while rebasing"):
		return "A rebase is in progress. Run `git rebase --continue` or cancel with `git rebase --abort`."
	case strings.Contains(msg, "cannot switch branch while reverting"):
		return "A revert is in progress. Run `git revert --continue` or cancel with `git revert --abort`."
	case strings.Contains(msg, "not fully merged"):
		return "The branch has commits that are not merged. Merge it first, or delete it with `git branch -D <name>` if you are sure."
	case strings.Contains(msg, "has no upstream branch"), strings.Contains(msg, "no upstream branch"):
		return "This branch has no upstream. Push it once with an upstream (Repora sets it automatically when you press P)."
	case strings.Contains(msg, "not possible to fast-forward"):
		return "The branch diverged from its upstream. Fetch, then integrate manually in a terminal."
	case strings.Contains(msg, "could not resolve head"):
		return "The repository has no commits yet. Create the first commit before switching branches."
	case strings.Contains(msg, "index.lock"), strings.Contains(msg, "another git process"):
		return "Another Git process is using this repository. Wait for it to finish, or remove the lock file only if you are sure it is stale."
	case strings.Contains(msg, "authentication failed"),
		strings.Contains(msg, "permission denied (publickey)"),
		strings.Contains(msg, "could not read username"),
		strings.Contains(msg, "terminal prompts disabled"):
		return "Authentication is required. Repora runs Git non-interactively; authenticate once in a terminal (SSH agent or credential helper)."
	case strings.Contains(msg, "nothing to commit"):
		return "There are no staged changes to commit."
	case strings.Contains(msg, "already exists"):
		return "That branch name already exists. Choose another name or check it out instead."
	default:
		return ""
	}
}
