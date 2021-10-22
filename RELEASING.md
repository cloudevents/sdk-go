# Releasing CloudEvents SDK for Golang

At the time of writing, there are two major versions being releases,

- v1.+ are released from the `release-1.y.z` branch.
- v2.+ are released from the `main` branch.

General rules that apply to both releases:

- Follow semver on API changes.
- Document what is in the release.

## Releasing v1.+

For v1 releases, just perform a GitHub release flow, and that is it.

## Releasing v2.+

_Note_: What is released is controlled by
[hack/tag-release.sh](./hack/tag-release.sh) script. Make sure the modules and
go modules are up-to-date. 

Using GitHub Actions,

Steps: 

1. Create a branch, in the form `release-v<major>.<minor>`, i.e:

   ```shell
   branch=release-v2.1
   git checkout -b $branch
   git push -u origin $branch
   ```

   Or using the GitHub UI: search for `release-v2.1` and then click create
   branch.

2. Update the release description.

That's it.

---

_UPDATE_: The following document is not required when using GitHub Actions. We
will keep the process documented for manual usage of the shell script or 100%
manual.

Releasing v2 is more tricky, the project has been broken into modules and are
released independently but there is a dependency chain.

_Note_: The following steps assume the repo is checked out directly. Switch to
`origin` to the remote name used if using a fork with a remote.

Manual Steps:

1. Create a branch, in the form `release-<major>.<minor>`, i.e:

   ```shell
   branch=release-2.1
   git checkout -b $branch
   git push -u origin $branch
   ```

1. Tag the new branch with the correct release tag, i.e:

   ```shell
   tag=v2.4.1
   git tag $tag
   git push origin $tag
   ```

1. Update and run `./hack/tag-release.sh` to update the dependencies.

   _Note:_ `./hack/tag-release.sh` has the tag config that needs to be updated.

   Then push the changes to the `release-x.y` branch.

   Or do it manually with the following:

   1. Update the each protocol to use the new release rather than the replace
      directive in their `go mod` files. Example for `stan`:

      ```shell
      tag=v2.1.0
      pushd protocol/stan/v2
      go mod edit -dropreplace github.com/cloudevents/sdk-go/v2
      go get -d github.com/cloudevents/sdk-go/v2@$tag
      popd
      ```

      _NOTE_: to reverse the `dropreplace` command, run
      `go mod edit -replace github.com/cloudevents/sdk-go/v2=../../../v2`

   1. Release each protocol with a new tag that lets go mod follow the tag to
      the source in the form `<path>/v<major>.<minor>.<patch>`, i.e.:

      ```shell
      tag=protocol/stan/v2.1.0
      pushd protocol/stan/v2
      git tag $tag
      git push origin $tag
      popd
      ```

1. Run `./hack/tag-release.sh --tag --push` to create a release of each
   sub-module.
   
   Or do it manually for each of the sub-modules with something like
   the following:

   ```shell
   tag=protocol/stan-v2.1.0
   git tag $tag
   git push origin $tag
   popd
   ```

1) Run `./hack/tag-release.sh --samples` to update the sample dependencies (both
   the core sdk and protocol) after all releases are published. Or do it manually
   with something like the following:

   ```shell
   tag=v2.1.0
   pushd sample/stan
   go mod edit -dropreplace github.com/cloudevents/sdk-go/protocol/stan/v2
   go get -d github.com/cloudevents/sdk-go/protocol/stan/v2@$tag
   go mod edit -dropreplace github.com/cloudevents/sdk-go/v2
   go get -d github.com/cloudevents/sdk-go/v2@$tag
   popd
   ```
