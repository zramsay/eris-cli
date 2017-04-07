---

type:   docs
layout: single
title: "Specifications | Assert Jobs Specification"

---

## Assert Jobs Specification

<div class="note">
  <em>Note: As of 2017, our product has been renamed from Eris to Monax. This documentation refers to an earlier version of the software prior to this name change (<= 0.16). Later versions of this documentation (=> 0.17) will change the <code>eris</code> command and <code>~/.eris</code> directory to <code>monax</code> and <code>~/.monax</code> respectively.</em>
</div>

Asserts can be used to compare two "things". These "things" may be the result of two jobs or the result against one job against a baseline. (Indeed, it could be the comparison of two baselines but that wouldn't really get folks anywhere).

Currently the following comparisons are supported:

```go
switch assertion.Relation {
case "==", "eq":
  if assertion.Key == assertion.Value {
    return assertPass()
  } else {
    return assertFail(assertion.Key, assertion.Value)
  }
case "!=", "ne":
  if assertion.Key != assertion.Value {
    return assertPass()
  } else {
    return assertFail(assertion.Key, assertion.Value)
  }
case ">", "gt":
  k, v, err := bulkConvert(assertion.Key, assertion.Value)
  if err != nil {
    return convFail()
  }
  if k > v {
    return assertPass()
  } else {
    return assertFail(assertion.Key, assertion.Value)
  }
case ">=", "ge":
  k, v, err := bulkConvert(assertion.Key, assertion.Value)
  if err != nil {
    return convFail()
  }
  if k >= v {
    return assertPass()
  } else {
    return assertFail(assertion.Key, assertion.Value)
  }
case "<", "lt":
  k, v, err := bulkConvert(assertion.Key, assertion.Value)
  if err != nil {
    return convFail()
  }
  if k < v {
    return assertPass()
  } else {
    return assertFail(assertion.Key, assertion.Value)
  }
case "<=", "le":
  k, v, err := bulkConvert(assertion.Key, assertion.Value)
  if err != nil {
    return convFail()
  }
  if k <= v {
    return assertPass()
  } else {
    return assertFail(assertion.Key, assertion.Value)
  }
}
```

Only number-like types may be compared using the `gt`, `ge`, `lt`, `le` notation.

Currently, the only way to call an assert value of equals for an array is to declare it as a string as such:

```
- name: assertBytesMemoryArray
  job:
    assert:
      key: $queryBytesMemoryArray
      relation: eq
      val: "[hello,marmots,how,are,you]"
```

**N.B.** -- yaml can be a bit testing. If you use the symbols notation make sure to put double quotes around the following:

* `"!="`
* `">="`
* `">"`


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Specifications](/docs/specs/)

