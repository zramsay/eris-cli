---

layout: single
title: "Tutorials | Solidity 3: Solidity Language Features"

---

# Solidity Series

This sequence of tutorials assumes that you have an understanding of the `eris` tooling to the point we ended in our [101 tutorial sequence](/docs/getting-started/).

This tutorial is part of our Solidity tutorial series:

* [The Five Types Model (Solidity 1)](solidity_1_the_five_types_model)
* [Action-Driven Architecture (Solidity 2)](solidity_2_action_driven_architecture)
* [Solidity Language Features (Solidity 3)](solidity_3_solidity_language_features)
* [Testing Solidity (Solidity 4)](solidity_4_testing_solidity)
* [Modular Solidity (Solidity 5)](solidity_5_modular_solidity)
* [Advanced Solidity Features (Solidity 6)](solidity_6_advanced_solidity_features)
* [Updating Solidity Contracts (Solidity 7)](solidity_7_updating_solidity_contracts)

# Introduction

In this tutorial we will cover some of the basic ones, types, interfaces, events, errors, and give a few examples of how these work in practice.

# Types - the basics

*Type-related information can be found in the official Solidity tutorial, under [types](https://github.com/ethereum/wiki/wiki/Solidity-Tutorial#types). It is good to have read that first.*

Solidity is a `statically typed language`, like for example C/C++ and Java. This may be new to people that have been working mostly with scripting/interpreted languages. What `statically typed` means is that when you declare a variable you must also include its type. For example: `myVar = 55;`is not allowed, but `int myInt = 55;` is. Types can be inferred by using `var` i.e. `var myVar = 55;` is allowed and will automatically get the type `uint8`. You must initialize a `var` when declaring it.

Types are checked at compile-time, so if you make a mistake you will get a compiler error. For example, this is not possible:

```javascript
contract Test {
  bool bVar;

  function causesError(address addr){
    bVar = addr;
  }
}
```

The error it would throw is this: `Error: Type address not implicitly convertible to expected type bool.`

## Type conversion

The compiler allows you to convert between types in certain cases. Let's say you have the number `1` stored in a `uint` variable, and you want to use it in another variable of type `int`. That is possible - but you generally have to do the conversion yourself. This is how you would do it:

```javascript

uint x = 1;
int y = int(x);

```

Type conversion is also checked at compile time and will generally be caught but there are exceptions; the most important one is when converting an address to a contract type. These type of casts can lead to bugs. We will be looking at some examples in a later section.

Finally, type conversion is something that should be used with care. It's good in some cases, but excessive and/or careless casting is usually a sign that the code is not well written and can sometimes have bad consequences (such as data-loss). Remember types are there for a reason.

# Contracts and Interfaces

Solidity uses the `contract` data-type to model smart contracts. It is very similar to a `class`.

A `contract` has a number of fields and methods; for example, the `contract` type can have a constructor, it can inherit from other contracts, etc.

The official tutorial has a number of simple [example contracts](https://github.com/ethereum/wiki/wiki/Solidity-Tutorial#simple-example) in it. Recently the Solidity designers have added `interface contracts`. These contracts allow functions to be abstract (have no body). Technically it has been possible to use "interface-ish" contracts before, but it has not been possible to make them truly abstract until now.

As I write this (2015-04-14) they haven't implemented all the features (as per the [story](https://www.pivotaltracker.com/story/show/88344782)), but it's practically good to go.

Here is an example. It is a simple interface with only one function in it.

```javascript
contract Depositor {
  function deposit(uint amount);
}
```

This function has no body, so cannot run on its own. Now we want to make a contract that is a `Depositor`, which means it implements this interface.

```javascript
contract HeyImADepositor is Depositor {

}
```

No, you're not. Why? You're not implementing the deposit function. If I try and compile this, it will fail. As I write this I don't yet get a compiler error, but the contract will not work (no bytecode). In order for the contract to work it has to create a function with the same signature as the deposit function but with a proper body.

```javascript
contract OkButNowIAm is Depositor {
  function deposit(uint amount) {}
}
```

Yes. Technically you are now a depositor, because you have implemented the `deposit` function as required by the Depositor interface.

Now we will make an interface that extends another interface but does not add to it.

```javascript
contract Depositoror is Depositor{}
```

Here is the implementation

```javascript
contract DepImpl is Depositoror {
  function deposit(uint amount) {}
}
```

The `DepImpl` contract will compile, and it will work exactly like `OkButNowIAm`. Next we're going to create an interface that extends two other interfaces.

```javascript
contract Depositor {
  function deposit(uint amount) returns (bool);
}

contract Withdrawer {
  function withdraw(uint amount) returns (bool);
}

contract BankUser is Depositor, Withdrawer {}
```

Now we implement `BankUser`, create a `Bank` interface, implement that and then and combine them.

```javascript
// Interface for banks.
contract Bank {
  // The return values would be to indicate that the transaction was successful.
  function makeDeposit(uint amount) constant returns (bool);
  function makeWithdrawal(uint amount) constant returns (bool);
}

// Dummy implementation of 'Bank'.
contract UBS is Bank {

  function makeDeposit(uint amount) returns (bool) {
    return true;
  }

  function makeWithdrawal(uint amount) returns (bool) {
    return true;
  }
}

// Implementation of BankUser
contract ABankUser is BankUser {

  Bank bank;

  function ABankUser(){
    bank = new UBS();
  }

  function deposit(uint amount){
    bank.makeDeposit(amount);
  }

  function withdraw(uint amount){
    bank.makeWithdrawal(amount);
  }
}
```

The `ABankUser` contract keeps a reference to a `Bank` contract to do the actual depositing. `Bank` is an interface, which means that any contract that implements that interface will do. In fact, we could make this contract even more generic by allowing the bank to be set. This is valid:

```javascript
contract ARiskyBankUser is BankUser {

  Bank bank;

  function deposit(uint amount){
    bank.makeDeposit(amount);
  }

  function withdraw(uint amount){
    bank.makeWithdrawal(amount);
  }

  function setBank(address addr){
    this.bank = Bank(addr);
  }
}
```

It has `risky` in it because it is not safe. First of all, `bank` starts out un-initialized, which means the `deposit` and `withdraw` functions might fail. Secondly, `setBank` has an address in the method signature and there is no guarantee that this contract is a bank. Finally, of course this is generally a bad contract because it has no permissions structure. It's just a demonstration of interfaces so it shouldn't have that, but it's still worth keeping in mind.

# Events

Events are used to dump information from Solidity contract code into the blockchain clients log. It is a way of making that information available to the "outside world". On top of the events themselves, most clients also have a way of capturing this output and encapsulating it in an event data-structures. This is particularly important for efficiency between the blockchain clients and the "outside world" which will rely upon these events in order for other things to happen.

Let us look at an example. We start by adding a new function to the BankUser interface:

```javascript
contract BankUser is Depositor, Withdrawer {
  function complain(bytes32 complaint);
}
```

Now we implement:

```javascript
contract ABankUser is BankUser {

  Bank bank;

  event Complain(address indexed userAddress, bytes32 indexed complaint);

  function ABankUser(){
    bank = new UBS();
  }

  function deposit(uint amount){
    bool result = bank.makeDeposit(amount);
    if(!result){
      complain("wtf");
    }
  }

  function withdraw(uint amount){
    bool result = bank.makeWithdrawal(amount);
    if(!result){
      complain("wtf");
    }
  }

  function complain(bytes32 complaint){
    Complain(msg.sender, complaint);
  }
}
```

What will happen here is that every time a `ABankUser` contract is executed, and the `complain` method is run, it will generate an event which can be read from the log. When using a client library like [Eris Contracts](https://www.npmjs.com/package/eris-contracts), you can set up a listener for this particular event. It is very simple. Assume that the contract for a particular `ABankUser` is named `bankUser123`. To generate a filter for that event we would simply do this:

```javascript
var filter = bankUser123.Complain();
```

Events are included in the contracts json ABI, and by calling the corresponding javascript function you get a filter. If we want to listen to and handle events continuously we could do this:

```javascript
filter.watch(callbackFun(data));

function callbackFun(data){
  var args = data.args;
  eMailTheManager(args.userAddress, args.complaint);
}
```

The `args` object will have fields that are named after the indexed fields in the contract, so you can decide when making the event what each of these fields should be called, and of course their types.

Regarding types: Contracts and "interfaces" are the same. There's no special interface type. The only difference is that an interface contract is allowed by the compiler to have abstract functions in it. Also, as we have seen, it is possible to coerce a contract into a super-contract i.e. there is no need to make an explicit cast:

```javascript
contract A {}

contract B is A {}

contract C {
  A a;

  function C(){
    a = new B();
  }
}
```

# Converting Addresses to Contracts

You can convert between contracts and addresses. This for example is allowed:

```javascript
function setB(address addr){
  b = B(addr);
}
```

There is no way of checking what type of contract is actually at that address though - or if it's even a contract. What this means is: **A contract can pass the compiler type checks but still be of the wrong type. Also, this is very hard to detect.**

Consider this:

```javascript
contract Greeter {
  function greet() returns (bytes32) {
    return "Hello!";
  }
}

contract Test {
  Greeter greeter;

  function setGreeter(address addr) {
    // Notice this is a cast, not instantiation.
    greeter = Greeter(addr);
  }

  function callGreeter() returns (bytes32) {
    return greeter.greet();
  }
}

contract Tester {
  Test t;
  bytes32 msg;

  function Tester(){
    t = new Test();
    Greeter greeter = new Greeter();
    t.setGreeter(address(greeter));
    msg = t.callGreeter();
  }
}
```

This compiles, and we can check to see that `Tester` has "Hello!" written into its `msg` field; however, if we change the constructor of `Tester` into passing its own address to `Test` - `t.setGreeter(address(this))` - it will still compile, but it will not function correctly. The reason is of course that the address we pass to `setGreeter` is not the address of a `Greeter`. When we call `Test.callGreeter`, the data will be properly formatted and a call will be made to the address of `Greeter`, but the receiving contract is a `Tester`, and thus it has no idea how to handle it.

One way of circumventing this is to only allow contracts to be added in very controlled ways, for example through dedicated factories or through methods that only certain accounts are allowed to access, but it is something that needs to be done with care.

# Errors

There is no real error handling system in Solidity (yet). There are no `try - catch` or `throw` statements, or something to that effect. Contract designers need to deal with errors themselves. Solidity does some sanity checks on arrays and such, but will often respond simply by executing the `(STOP)` instruction. According to the developers, this is just put in as a placeholder until a more sophisticated error handling and recovery system is put in place.
