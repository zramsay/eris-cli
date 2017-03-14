# Variables Specification

Variables can be used for nearly every eris [jobs](/docs/specs/jobs_specification) field (largely with the exception of nonce and wait).

eris:jobs variables will always begin with a dollar sign `$`. This is what will trigger the variable expansion.

Variables come in the following types:

* [job result variables](#jobResultVars)
* [set variables](#setVars)
* [reserved variables](#reservedVars)
* [fallback function](#fallBack)
* [tuple returns](#tupleReturns)
* [array packing and returns](#arrays)

## <a name="jobResultVars"></a>Job Result Variable

The result of every job is set as a variable with the `JobName` as the key and the `JobResult` as the value. The `JobResult` for transaction jobs is the transaction hash. The `JobResult`  for contract deployments is the address of the contract. The `JobResult` for queries and calls is the return value from the blockchain or the query.

The `JobResults` which are able to be retrieved from query functions will vary and depend largely on the fields which are returnable from eris-client's tooling.

## <a name="setVars"></a>Set Variables

Set variables will take the `JobName` and use the `val` field from the job file to set the variable.

## <a name="setVars"></a>Variable Types

If you're using solidity then you will be familiar with variable types. Here is how eris:pm deals with variable types:

* `address` - addresses should be given according to the 40 character string **without** the leading `0x`
  * Example: 1040E6521541DAB4E7EE57F21226DD17CE9F0FB7
* `int` && `uint` -- integers (signed and unsigned) should be given according to their plain text rendering of the digits
  * Example: 99999
* `bool` - `true` or `false`
* `string` and `byteX` -- just give it a string
  * Example: marmatoshi

For a more complete handling of the types, please see the epm.yaml in tests/fixtures/app06 directory of the repository.

## <a name="reservedVars"></a>Reserved Variables

The following are reserved variables:

* `$block`: will return a string version of the current block height
* `$block+X`: where `X` can be any digit; will return a string version of the current block height `+X` blocks
* `$block-X`: where `X` can be any digit; will return a string version of the current block heigh `-X` blocks

## <a name="fallBack"></a>Fallback Function

In order to test your fallback function in your contract using the call job, simply put the name of your function as "()" and the fallback function will be called. See app37.

## <a name="tupleReturns"></a>Tuples and Returns

eris:jobs can now effectively handle multiple return values for all static types such as

* `address` `int` `uint` `bool` `bytes(1-32)`

You can access these in your jobs by specifying the name of the value returned. If you have not appended a name to the value returned, simply call them by the order in which they are returned. For example:

```
contract tuples {
// for a job $getBools we could call this by
// $getBools.0 to get true and $getBools.1 to get false
    function getBools() returns (bool, bool) { return (true, false); }
// for a job $getInts here we would call $getInts.a == 3,$getInts.b == 5
    function getInts() returns (uint a, int b) { return (3, 5) }
}
```

for now the epm cannot handle dynamic types such as

* `string` `bytes` `struct`

Hold with us while the marmots get those in control :)

## <a name="arrays"></a> Array Packing and Returns

eris:jobs can now handle packing and returning of arrays with some caveats. In order to pack an array value in, you must declare it inside square brackets. For an example, see [app31](https://github.com/eris-ltd/eris/tree/master/tests/fixtures/app31/epm.yaml). Until then, you can declare arrays for most static types such as:

*  `int` `uint` `bool` `bytes(1-32)`

We currently do not handle packing of 2D arrays nor arrays of `address`, `string`, `bytes`, or `struct`. These are scheduled for upcoming releases.
