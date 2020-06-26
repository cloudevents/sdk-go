# Releasing CloudEvents SDK for Golang

At the time of writing, there are two major versions being releases,

- v1.+ are released from the `release-1.y.z` branch.
- v2.+ are released from the `master` branch.

General rules that apply to both releases:

- Follow semver on API changes.
- Document what is in the release.

## Releasing v1.+

For v1 releases, just perform a GitHub release flow, and that is it. 

## Releasing v2.+

Releasing v2 is more tricky, the project has been broken into modules and are released independently but there is a depency chain.

_Note_: Any tag that matches `v*` or `protocol*` will produce a draft GitHub release using GitHub Actions.

_Note_: The following steps assume the repo is checked out directly. Switch to `origin` to the remote name used if using a fork with a remote.

Steps:

1. Create a branch, in the form `release-<major>.<minor>`, i.e:
     ```
     branch=release-2.1
     git checkout -b $branch
     git push -u origin $branch
     ```
1. Tag the new branch with the correct release tag,
    ```
    tag=v2.1.0
    git tag $tag
    git push origin $tag
    ```