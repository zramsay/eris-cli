---

layout: single
title: "Tutorials | Solidity 7: Updating Solidity Contracts"

---

## Solidity Series

<div class="note">
{{% data_sites rename_docs %}}
</div>

This sequence of tutorials assumes that you have an understanding of the `monax` tooling to the point we ended in our [101 tutorial sequence](/docs/getting-started/).

This tutorial is part of our Solidity tutorial series:

* [Part 1: The Five Types Model](/docs/solidity/solidity_1_the_five_types_model)
* [Part 2: Action-Driven Architecture](/docs/solidity/solidity_2_action_driven_architecture)
* [Part 3: Solidity Language Features](/docs/solidity/solidity_3_solidity_language_features)
* [Part 4: Testing Solidity](/docs/solidity/solidity_4_testing_solidity)
* [Part 5: Modular Solidity](/docs/solidity/solidity_5_modular_solidity)
* [Part 6: Advanced Solidity Features](/docs/solidity/solidity_6_advanced_solidity_features)
* [Part 7: Updating Solidity Contracts](/docs/solidity/solidity_7_updating_solidity_contracts)

## Introduction

This tutorial series looks at modular systems of smart-contracts, and how to continuously update the code in a reliable way. Most contracts in your application will become obsolete at some point, and will require an update. Same as in other applications. It could be because new features must be added, a bug is found, or because a better, more optimized version has been made. Updating could of course cause problems so it must be done with care. Some of the things one must ensure is that:

- updating is possible.
- the new contract works as intended.
- all the calls made during the replacement procedure was executed successfully.
- replacing the contract has no side-effects in other parts of the system.

The first point may seem obvious but it usually requires a lot of work, because updating is not possible by default; the reason is because of how accounts, code and storage works.

## Accounts, Code and Storage

A very important property of EVM contracts is that when a contract has been uploaded to the chain, the code can never be changed. Contracts are stored in special account objects, and these object has references to the contract (byte) code, and a database, and some other things. The database is a key-value store, also known as 'storage', and is where data such as the values of contract fields is stored.

When contracts are created, a new account is first made, then the code is loaded into a VM which runs the constructor part, initializes fields etc., and then adds the runtime portion (or body) of the contract to the account. After that is done, there is no way to change the code, and there is no way to update the database except through that code.

But what if you want to change the code? What if a bug is discovered?

The way you solve that is by connecting several contracts. Contract `C` could call contract `D` as part of its functionality, and the address to `D` could be settable in `C`, meaning it would be possible to change what `D` is. This is best explained through a series of simple examples.

### A simple storage contract

```javascript
contract Data {

    uint public data;

    function addData(uint data_) {
        if(msg.sender == 0x692a70d2e424a56d2c6c27aa97d1a86395877b3a)
            data = data_;
    }

}
```

This simple contract allows a user to add and read an unsigned integer. The only account that is allowed to add data is the account with address `0x692a...`. This address is a hex literal, so is added to the bytecode when the contract is compiled.

A potential problem is that we might want to replace this address later, or even the entire validation procedurer, but we can't because of how code and storage works. A simple way of making the contract more flexible is to store the current owner address in storage instead, and make it possible to change.

```javascript
contract DataOwnerSettable {

    uint public data;

    address public owner = msg.sender;

    function addData(uint data_) {
        if(msg.sender == owner)
            data = data_;
    }

    function setOwner(address owner_) {
        if(msg.sender == owner)
            owner = owner_;
    }

}
```

This contract has an `owner` field (mapped to storage). It is initialized with the address of the account that creates the contract, and can later be changed by the current owner by calling `setOwner`. The guard inside `addData` is still the same; the only thing that changed is that the owner address is no longer hard-coded.

## Delegation

What if a settable owner is not enough, though? What if we want to be able to update not only the owner address, but the entire validation process? That is possible. We will do it in two steps. First we move the account validation code into a different contract.

```javascript
contract AccountValidator {

    address public owner = msg.sender;

    function validate(address addr) constant returns (bool) {
        return addr == owner;
    }

    function setOwner(address owner_) {
        if(msg.sender == owner)
           owner = owner_;
    }

}

contract DataExternalValidation {

    uint public data;

    AccountValidator _validator;

    function DataExternalValidation(address validator) {
        _validator = AccountValidator(validator);
    }

    function addData(uint data_) {
        if(_validator.validate(msg.sender))
            data = data_;
    }

    function setValidator(address validator) {
        if(_validator.validate(msg.sender))
            _validator = AccountValidator(validator);
    }
}
```

To use this, we first create an `AccountValidator` contract; it has the `owner` field now, and that field is automatically initialized with an account address. Then we create a `DataExternalValidation`-contract and inject the address of the validator through the contract constructor. When someone tries to write to `data`, it will call the `validate` function of the current validator contract to do the check rather then storing (or hard coding) the `owner` address and doing the equality check internally. Everything that has to do with access control is now delegated to the validator contract.

This is very nice, because it is now possible to replace the contract that does the actual check. Not only does it decouple this from the data, but since the `AccountValidator` is its own contract, we could potentially use that contract in other contracts as well and thus give `owner` control over more contracts then just one.

One thing remains though. We still can't replace the code! All we have done is move the validation code out of the contract. The code of the `AccountValidator` contract can't be changed anymore then that of the data contract. Fortunately, Solidity provides a very simple and powerful workaround - abstract functions.

## Abstract Functions

Using abstract functions, the validator contract could be changed into this:

```javascript
contract AccountValidator {
    function validate(address addr) constant returns (bool);
}


contract SingleAccountValidator is AccountValidator {

    address public owner = msg.sender;

    function validate(address addr) constant returns (bool) {
        return addr == owner;
    }

    function setOwner(address owner_) {
        if(msg.sender == owner)
            owner = owner_;
    }

}
```

With these contracts, the data contract no longer works with a concrete validator contract, but an abstract (interface) representation. This makes sense, because it does not really needs to know what the `validate` function actually does, it only needs to know the signature.

Interfaces works the same way as it does in most other object-oriented languages, just declare functions without a body and they become abstract.

We still can't change the code stored in a contract account, but we can change the code that is executed when a function is called, by delegating some functionality to other contract which are allowed to be replaced; all we need to do is change the validator contract to a different contract. For example, if we want to allow more owners then one we could use an instance of this contract:

```javascript
contract MultiAccountValidator is AccountValidator {

    mapping(address => bool) public owners;

    function MultiAccountValidator() {
        owners[msg.sender] = true;
    }

    function validate(address addr) constant returns (bool) {
        return owners[addr];
    }

    function addOwner(address addr) {
        if(owners[msg.sender])
            owners[addr] = true;
    }
}
```

## Summary

Proper delegation is an important part of smart-contract systems. It is also something one has to consider from the very start, because the rules for how a set of contracts can be updated is generally contained in the contracts themselves. Also, the more contracts that are in the system the harder they become to manage, and a strategy that makes a small system work may not be suitable for a medium-sized or large one.

Another thing to keep in mind is that modularity comes with a cost, because it requires more code, storage variables and calls. On the public chain, where the gas limitations are quite severe (for obvious reasons), even a small modular system could be hard to deploy and run. Generally, when it comes to scalability vs. efficiency I tend to go with scalability. The large, expensive contracts in an excessively modular system can after all be improved and replaced, but if the contracts are locked down that may not be an option.

In our opinion, it is very important to at least acknowledge that the code is going to need updates, and at some point there must be a good policy for how it can be done. The alternative is to not have a plan and fail. And then maybe fail again, and again, until eventually it becomes clear.


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Solidity Tutorials](/docs/solidity/)
