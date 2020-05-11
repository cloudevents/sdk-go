# Contributing to CloudEvents Go SDK

:+1::tada: First off, thanks for taking the time to contribute! :tada::+1:

We welcome contributions from the community! Please take some time to become
acquainted with the process before submitting a pull request. There are just
a few things to keep in mind.

## Pull Requests

Typically a pull request should relate to an existing issue. If you have
found a bug, want to add an improvement, or suggest an API change, please
create an issue before proceeding with a pull request. For very minor changes
such as typos in the documentation this isn't really necessary.

For step by step help with managing your pull request, have a look at our
[PR Guidelines](pr_guidelines.md) document.

### Sign your work

Each PR must be signed. Be sure your `git` `user.name` and `user.email` are configured
then use the `--signoff` flag for your commits.

```console
git commit --signoff
```

### Style Guide

Code style for this module is maintained using [`eslint`](https://www.npmjs.com/package/eslint).
When you run tests with `npm test` linting is performed first. If you want to
check your code style for linting errors without running tests, you can just
run `npm run lint`. If there are errors, you can usually fix them automatically
by running `npm run fix`.

Linting rules are declared in [.eslintrc](https://github.com/cloudevents/sdk-javascript/blob/master/.eslintrc).
