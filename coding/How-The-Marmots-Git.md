We follow an almost identical flow as described [in this post that has become a base go-to for git product lifecycle](http://nvie.com/posts/a-successful-git-branching-model/)

In general, we use the following model:

* `master` branch encapsulates our current release. **The only time you should merge to master is when you are releasing**. Note, that the previous sentence applies to a "major" repository. If you are working in a repository which is basically a library the above default rule need not apply. If you are unsure what repository you are working in (namely, whether it is a "major" repository or a "library" repository) then **ask**. If master goes red (namely, its tests fail) **stop what you are doing and make it go green**; before you do anything else. 
* `develop` branch encapsulates our upcoming release. The version/version.go (or equivalent) on develop branch should, in general, be whatever `master` is on plus one. Develop branch is our "working, semi-stable" branch. It should be nearly always green. Sometimes develop can be red, but only if there is something that is being worked on. Always, always, always, though, it should be green before merging into master.
* `featureBranch` are new features or fixes which we are working on. These should be branched from `develop` and pull requested into `develop`. 

Pull requests do's and dont's:

* `do` document your pull request. Simply using your most recent merge comment is OK if it is a dead simple PR, otherwise, you should detail what the PR is actually doing.
* `do` PR into `develop`
* `do not` PR into `master` (unless you are a maintainer of a repository and/or are releasing)
* `do` please PR from eris-ltd organization push. In general we do not recommend eris folks first fork an eris repo and PR from their personal fork into the eris-ltd repository. The reason here is that most of our repositories use env vars and other details which will only be populated in eris-ltd repositories. In general, for most of our repositories, personal PRs will fail the CI system. If you are a marmot, then please push your work to a featureBranch on eris-ltd repository and then PR that into develop. 
* `do not` merge your own PR's. When we were small this happened a lot. As our team grows, it is increasingly important to have a second eye on our work. If you merge your own PR into develop, then at least get a :+1: from a colleague who may be affected by the feature you are adding.
* `do` merge feature branches back to *develop* using a squash merge, unless you have specific reasons not to. This ensures a cleaner history in *develop* and easier trouble-shooting in case of problems as feature contributions from different team members are represented as single commits. **Note**: Upon attempting to delete the local branch after using a squash merge Git will issue a warning that the local branch was not fully merged. These branches can be deleted forcefully via git branch -D <branch>