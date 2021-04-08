# Maintainer's Guide

## Tips

Here are a few tips for repository maintainers.

* Stay on top of your pull requests. PRs that languish for too long can become difficult to merge.
* Work from your own fork. As you are making contributions to the project, you should be working from your own fork just as outside contributors do. This keeps the branches in github to a minimum and reduces unnecessary CI runs.
* Try to proactively label issues with backport labels if it's obvious that a change should be backported to previous releases.
* When landing pull requests, if there is more than one commit, try to squash into a single commit. Usually this can just be done with the GitHub UI when merging the PR. Use "Squash and merge".
* Triage issues once in a while in order to keep the repository alive. During the triage:
  * If some issues are stale for too long because they are no longer valid/relevant or because the discussion reached no significant action items to perform, close them and invite the users to reopen if they need it.
  * If some PRs are no longer valid but still needed, ask the user to rebase them
  * If some issues and PRs are still relevant, use labels to help organize tasks
  * If you find an issue that you want to create a fix for and submit a pull request, be sure to assign it to yourself so that others maintainers don't start working on it at the same time.

## Branch Management

The `main` branch is is the bleeding edge. New major versions of the module
are cut from this branch and tagged. If you intend to submit a pull request
you should use `main HEAD` as your starting point.

Each major release will result in a new branch and tag. For example, the
release of version 1.0.0 of the project results in a `v1.0.0` tag on the
release commit, and a new branch `release-1.y.z` for subsequent minor and patch
level releases of that major version if necessary. However, development will continue
apace on `main` for the next major version - e.g. 2.0.0. Version branches
are only created for each major version. Minor and patch level releases
are simply tagged.

