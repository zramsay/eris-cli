# The Monax Platform
- [ ] Agree on a version to which the #develop branch will be bumped, for the _next_ release.
- [ ] execute keys's release checklist
- [ ] execute db's release checklist
- [ ] execute compiler's release checklist
- [ ] execute release_script checklist
- [ ] draft a [PR](#draft-a-release-pr) from cli:develop to cli:master on Github.
- [ ] when tests are stable, merge the PR to master
- [ ] [tag](#tag-the-release) the release with a changelog
- [ ] execute `misc/release/release.sh` while on the #master branch. This will:
  * publish a release to github
  * add the changelog from the most recent tag to github
  * cross compile the cli for various platforms
  * builds .deb and .rpm packages and creates APT and YUM repositories
  * upload files to Amazon S3
- [ ] version bump develop
- [ ] update the `brew` formula (happens semi-automatically when new release is tagged)


# keys
- [ ] [draft](#draft-a-release-pr) a PR from develop to master
- [ ] merge develop to master
- [ ] [tag](#tag-the-release) the release with a changelog
- [ ] once tests pass ensure images pushed
- [ ] version bump develop

# db
- [ ] [draft](#draft-a-release-pr) a PR from develop to master
- [ ] merge develop to master
- [ ] [tag](#tag-the-release) the release with a changelog
- [ ] once tests pass ensure images pushed
- [ ] version bump develop

# compilers
- [ ] [draft](#draft-a-release-pr) a PR from develop to master
- [ ] merge develop to master
- [ ] [tag](#tag-the-release) the release with a changelog
- [ ] once tests pass ensure images pushed
- [ ] version bump develop
- [ ] start compilers remote service

# cli's release_script
- [ ] install `go get github.com/aktau/github-release`
- [ ] `export GITHUB_TOKEN=...`. pick that from the Github account
- [ ] `export AWS_ACCESS_KEY=...` and `export AWS_SECRET_ACCESS_KEY=...` to point to Monax's Amazon account.
- [ ] acquire `linux-private-key.asc` and `linux-private-key.asc` files with GPG keys for signing .deb and .rpm packages or generate them:

  ```
  gpg2 --gen-key
  gpg2 --export-secret-keys -a KEYID > linux-private-key.asc
  gpg2 --export -a KEYID > linux-public-key.asc
  ```
**WARNING:** Be careful not to commit the gpg keys

# Draft a release PR
**note:** for release PRs from #develop to #master use the standard PR title: `v0.12.0 Release` and include the changelog in the PR description. This applies to all repo's included as part of a release. An example: ![Release PR](http://i.imgur.com/IAm5pdN.jpg)

# Tag the release 
While on #master, run the `git tag -a vX.XX.XX -m 'ChangeLog: ...'` command and paste the ChangeLog text between single quotes; use the ChangeLog of this form (no spaces before `*` and the word `ChangeLog` at line beginnings):

  ```
  ChangeLog:
  * a lot of improvements to documentation and README files.
  * container number flag `-n N` is removed as an unused feature
  * many small fixes and improvements for easier integration testing
  ...
  ```
Run the `git push origin --tags` command (still while on #master) to publish the tag.
