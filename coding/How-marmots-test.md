Testing `eris` is a challenge. 

## Notes (will clean up later)

* `deploy-*` should consume artifacts from `build-*` as needed
* `test-*` should consume artifacts from `build-*` as needed
* `test-*` should NOT publish artifacts
* `deploy-*` should PUSH to EITHER quay.io (according to the rules herein) OR to binaries_cache (s3?)
* `deploy-*` should NOT publish artifacts

## Builds to keep (remember to also remove artifacts under advanced)

* `*-featurebranches` - 5 builds
* `*-develop` - 10 builds
* `*-master` - 10 builds