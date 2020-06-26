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

1. Tag the new branch with the correct release tag, i.e:
    ```
    tag=v2.1.0
    git tag $tag
    git push origin $tag
    ```
1. Run `./hack/tag-release.sh --tag --push` to update and tag the dependencies. Or do it manualy with the following:
   
  1. Update the each protocol to use the new release rather than the replace directive in their `go mod` files. Example for `stan`:
    ```
    tag=v2.1.0
    pushd protocol/stan/v2
    go mod edit -dropreplace github.com/cloudevents/sdk-go/v2
    go get -d github.com/cloudevents/sdk-go/v2@$tag
    popd
    ```
    _NOTE_: to reverse the `dropreplace` command, run `go mod edit -replace github.com/cloudevents/sdk-go/v2=../../../v2`

  1. Release each protocol with a new tag that lets go mod follow the tag to the source in the form `<path>/v<major>.<minor>.<patch>`, i.e.:
    ```
    tag=protocol/stan/v2.1.0
    pushd protocol/stan/v2
    git tag $tag
    git push origin $tag
    popd
    ```

1. Run `./hack/tag-release.sh --samples` to update the sample dependencies (both the core sdk and protcol) after all releases are published. Or do it manualy with the following:
    ```
    tag=v2.1.0
    pushd sample/stan
    go mod edit -dropreplace github.com/cloudevents/sdk-go/protocol/stan/v2
    go get -d github.com/cloudevents/sdk-go/protocol/stan/v2@$tag
    go mod edit -dropreplace github.com/cloudevents/sdk-go/v2
    go get -d github.com/cloudevents/sdk-go/v2@$tag
    popd
    ```
