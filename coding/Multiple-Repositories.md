At eris we make a lot of "unix-y" style tools which tend to do a single purpose thing. These tools are generally then packaged into higher level tooling until they are eventually "exposed" to the eris-cli (at some level of abstraction).

But what this means is that we have a lot of repositories which developers are expected to be able to work with. This is not a simple matter. 

eris-cli needs to stay in sync with: eris-db, eris-cm, eris-pm (although the tightness of the coupling is still a WIP). This mainly impacts on how we test.

We also have a decent amount of vendored dependencies. For example, eris-pm wraps mint-client, eris-compilers, eris-keys, and eris-abi. And when you work with the "lower" level packages you still need to manage these via vendoring.

## Testing in Multiple Repository Environments

How we manage our testing our stack is predominantly around env vars which are kept circle.yml (or equivalent CI runner). 

All test code should respect these global env vars. The global env vars should be in the form `ERIS_PM_BRANCH`. This makes for a single place where developers can modify the branch of an external repository which will generally be cloned and used. 

*Example* -- https://github.com/eris-ltd/eris-cli/blob/master/tests/test_stack.sh#L48-L56

## Working With Go Vendored Dependencies

Godeps is (currently) used to handle vendored dependencies for eris tooling. It is a bit of a bear to work with but is manageable if used carefully.

Definitions: 

For the purposes of this section **target repo** is the repository which is vendoring lower level tooling and **source repo** is the repository which is being vendored.

Workflow:

- [ ] `cd sourceRepo && git checkout -b newFeature`
- [ ] update source repo tests
- [ ] update source repo code
- [ ] ensure source repo tests green
- [ ] `cd targetRepo && git checkout -b newFeature && godep update github.com/eris-ltd/sourceRepo/...`
- [ ] update target repo tests
- [ ] update target repo code
- [ ] ensure target repo tests green
- [ ] `cd sourceRepo && git push`
- [ ] make sure sourceRepo CI goes green
- [ ] `cd targetRepo && git push`
- [ ] make sure targetRepo CI goes green

Repeat as necessary.

Throw your computer out the window as necessary.