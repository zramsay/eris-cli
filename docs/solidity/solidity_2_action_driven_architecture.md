---

layout: single
title: "Tutorials | Solidity 2: An Action-Driven Architecture"

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

The system proposed in part 1 is a good system in theory. It has good separation of concerns, is very modular, and is set up to handle permissions. This is how a typical system would look:

{{< image src="/images/docs/SSoSC2-1.png" >}}

The way contracts interact with each other is through the Doug contract. This is an illustration of a normal sequence of calls:

{{< image src="/images/docs/SSoSC2-2.png" >}}

There is a big problem with these type of systems however. Let's say we have a system with 10 controllers and 10 databases. Each controller has 4 functions in their public API on average. What would this mean? It means we need to add 40 (!) functions to the ALC (Application logic contract) in order to access them all. And that's not enough. Every time we update the system and add new controllers, or modify the exiting ones, we'll have to swap out the entire ALC!

We could try and mitigate this by dividing the ALC up into multiple contracts. We could also omit the ALC entirely and instead add multiple controllers that we call directly. Both of these solutions are perfectly fine, but we will no longer have a single entry-point. Moreover, we still have to replace entire contracts in order to do minor changes to the logic.

Action driven architecture:

- Good, flexible permissions management. One unit of functionality - one action - one permission.  contract is not a good divider, because removing the contract and performing one of its basic functions (like depositing into an account) would normally require very different permission levels. Actions are better.

- Unified interface - all actions work the same way. One pipeline. All action-driven systems are written the same way. Makes auto generation of UI widgets and other things easy, and similarity between systems makes the code easy to read as well.

- Fully extendable. Individual actions can be added and removed without affecting the rest of the system.

- Expensive, as there is lot of code and contracts, so usually not good in production environments.

## Actions

PRODOUG was made secure by encapsulating all the code that users could run in something called actions. All incoming transactions had to be on a specific format, and had to be sent through the 'action manager' contract (with a few exceptions). Here's an example of a super simple action interface:

```javascript
contract SomeAction {
  function execute(type1 par1, type2 par2, ....) constant returns (bool result) {}
}
```

The way you manage and call action contracts is by keeping references to them in a manager contract:

```javascript
// The action manager
contract ActionManager {

  // This is where we keep all the actions.
  mapping (bytes32 => address) actions;

  function execute(bytes32 actionName, bytes data) returns (bool) {
    address actn = actions[actionName];
    // If no action with the given name exists - cancel.
    if (actn == 0x0){
      return false;
    }
    // No type conversion possible here, for now.
    actn.call(data);
    return true;
  }

  // Add a new action.
  function addAction(bytes32 name, address addr) {
    actions[name] = addr;
  }

  // Remove an action.
  function removeAction(bytes32 name) constant returns (bool) {
    if (actions[name] == 0x0){
      return false;
    }
    actions[name] = 0x0;
    return true;
  }

}
```

**Important**

Since we must allow generic arguments, we must pass something into the action that can stand for any number of arguments of any type - like an `Object` in java, an `interface{}` in Go, or a `*void` in C. This is not fully supported in Solidity, but the first thing that that will be useful in this case is probably going to be byte arrays - which is basically how this worked in LLL. Byte arrays are fully generic, so what we'd do here (for now) is to use the [deprecated javascript library](https://github.com/monax/legacy-contracts.js) which makes it very simple to convert arguments into properly formatted call-data.

**Controller and database**

An action driven system does everything through action contracts. They can contain any logic, but normally they would be fairly small and focus on one or a few things. They have access to Doug, just like normal contracts, but validation is not done through Doug but through the action manager.

If you have read part 1, you'll notice notice that we're cheating here by adding the actions map to the action manager itself, which is wrong. The final contract version will keep it in a database contract.

This is a diagram over how the calls would look in a simple fund manager where you can add and remove users, and make deposits and withdrawals with the bank.

{{< image src="/images/docs/SSoSC2-3.png" >}}

There is of course no real security yet. At this point we just have a simple action system. People can add actions to it, remove them, and execute them. Before we can add any actions to it we have to add another component - the Doug. Even though the action manager is technically part CMC (contract managing contract), we need a Doug as well. It will link the actions and action manager with the other contracts in the system, such as databases. We'll start with a namereg type Doug similar to the one in part 1.

```javascript
// The Doug contract.
contract Doug {

    address owner;

    // This is where we keep all the contracts.
    mapping (bytes32 => address) public contracts;

    // Constructor
    function Doug(){
        owner = msg.sender;
    }

    // Add a new contract to Doug. This will overwrite an existing contract.
    function addContract(bytes32 name, address addr) returns (bool result) {
        if(msg.sender != owner){
            return false;
        }
        DougEnabled de = DougEnabled(addr);
        // Don't add the contract if this does not work.
        if(!de.setDougAddress(address(this))) {
            return false;
        }
        contracts[name] = addr;
        return true;
    }

    // Remove a contract from Doug. We could also selfdestruct if we want to.
    function removeContract(bytes32 name) returns (bool result) {
       address cName = contracts[name];
        if (cName == 0x0){
            return false;
        }
        if(msg.sender != owner){
            return false;
        }
        // Kill any contracts we remove, for now.
        DougEnabled(cName).remove();
        contracts[name] = 0x0;
        return true;
    }

    function remove(){
        if(msg.sender == owner){
            selfdestruct(owner);
        }
    }

}
```

We will also add a super simple bank, or credit contract.

```javascript
// The Bank contract
contract Bank {

  // This is where we keep all the permissions.
  mapping (address => uint) public balances;

  // Endow an address with coins.
  function endow(address addr, uint amount) returns (bool) {
    balances[addr] += amount;
  }

  // Charge an account 'amount' number of coins.
  function charge(address addr, uint amount) returns (bool){
    if (balances[addr] < amount){
      // Bounces if balance is lower then the amount.
      return false;
    }
    balances[addr] -= amount;
    return true;
  }

}
```

This is how the system would be initialized:

1) Deploy the Doug contract.
2) Deploy the action manager contract and register it with the Doug contract under the name "actions".
3) Deploy the bank contract and register it with the Doug contract under the name "bank".

What we need do next is to add an action for endowing an address with coins, and one for charging it. We need to add one more function to the actions interface though - the setDougAddress function. This function is what will give actions (indirect) access to all the contracts in the system so they can carry out their work. It is also an important security measure. We will use the DougEnabled contract from part 1.

```javascript
contract DougEnabled {
    address DOUG;

    function setDougAddress(address dougAddr) returns (bool result){
        // Once the doug address is set, don't allow it to be set again, except by the
        // doug contract itself.
        if(DOUG != 0x0 && dougAddr != DOUG){
            return false;
        }
        DOUG = dougAddr;
        return true;
    }

    // Makes it so that Doug is the only contract that may kill it.
    function remove(){
        if(msg.sender == DOUG){
            selfdestruct(DOUG);
        }
    }

}
```

The basic action template starts like this:

```javascript
contract Action is DougEnabled {}
```

Note that we don't include an 'execute' function for the reasons mentioned above. We will add the execute function on a per-action basis.

To lock these contracts down, we only allow the contract currently registered as `actions` to call the functions. Much like the `FundManagerEnabled` contract in part 1.

```javascript
contract ContractProvider {
  function contracts(bytes32 name) returns (address){}
}

contract ActionManagerEnabled is DougEnabled {
  // Makes it easier to check that action manager is the caller.
  function isActionManager() internal constant returns (bool) {
    if(DOUG != 0x0){
      address am = ContractProvider(DOUG).contracts("actions");
      if (msg.sender == am){
            return true;
      }
    }
    return false;
  }
}
```

The new action base class is this:

```javascript
contract Action is ActionManagerEnabled {}
```

Here's the endow action contract.

```javascript
// The Bank contract (the "sub interface" we need).
contract Endower {
  function endow(address addr, uint amount) {}
}

// The endow action.
contract ActionEndow is Action {

  function execute(address addr, uint amount) returns (bool) {
    if(!isActionManager()){
      return false;
    }
    ContractProvider dg = ContractProvider(DOUG);
    address endower = dg.getContract("bank");
    if(endower == 0x0){
      return false;
    }
    Endower(endower).endow(addr, amount);
    return true;
  }
}
```

This is the action for charging.

```javascript
// The Bank contract (or the "sub interface" we need).
contract Charger {
  function charge(address addr, uint amount) returns (bool) {}
}

// The charge action.
contract ActionCharge {

  function execute(address addr, uint amount) returns (bool) {
    if(!isActionManager()){
      return false;
    }
    ContractProvider dg = ContractProvider(DOUG);
    address charger = dg.getContract("bank");
    if(charger == 0x0){
      return false;
    }
    Charger(charger).charge(addr,amount);
    return true;
  }

}
```

When we add these actions to the action manager it will be possible for users to execute them and work with the bank contract that way. Note that it is still possible to interact with the bank contract directly so the actions are not useful yet, but we will fix that.

## Permissions

Step 2 is to control Doug contract access from the outside. It should only be possible to interact with the contracts through actions, and it should only be possible to run actions through the actions manager. The first thing we need to do is make sure the action manager (and later the action database) calls the 'setDougAddress' function that we added to the actions. It should call it and pass the DOUG address to the action as soon as it's registered. If the function returns false, that means it already has a doug address set which in turn means the action should not be registered with the action manager at all. It is unsafe.

We also need to add the DOUG address to the action manager. In fact, the bank and all other contracts like it should have a function that allows the DOUG value to be set so we'll just doug-enable all of them.

The action manager will also get a 'validate' function that can be called by other contracts to ensure that only actions can call them, and we will also break out the actions list into a separate actions database contract so that we can modify the action manager without having to clear all the actions.

### The Updated Contracts

This is the new action manager. We're adding the setDougAddress functionality when adding actions and also an 'active contract' field that will be used for validation. We will make `ActionDb`callable only from the action manager now, but there will be an even better system later.

```javascript
contract ActionDb is ActionManagerEnabled {

  // This is where we keep all the actions.
  mapping (bytes32 => address) public actions;

  function addAction(bytes32 name, address addr) returns (bool) {
    if(!isActionManager()){
      return false;
    }
    actions[name] = addr;
    return true;
  }

  function removeAction(bytes32 name) returns (bool) {
    if(!isActionManager()){
      return false;
    }
    if (actions[name] == 0x0){
      return false;
    }
    actions[name] = 0x0;
    return true;
  }

}
```

```javascript
// The new action manager.
contract ActionManager is DougEnabled {

  // This is where we keep the "active action".
  address activeAction;

  function ActionManager(){
  }

  function execute(bytes32 actionName, bytes data) returns (bool) {
    address actionDb = ContractProvider(DOUG).contracts("actiondb");
    if (actionDb == 0x0){
      return false;
    }
    address actn = ActionDb(actionDb).actions(actionName);
    // If no action with the given name exists - cancel.
    if (actn == 0x0){
      return false;
    }
    // Set this as the currently active action.
    activeAction = actn;
    // Run the action. Any contract that calls 'validate' now will only get 'true' if the
    // calling contract is 'actn'. Again - no return value check (true/false).
    actn.call(data);
    // Now clear it.
    activeAction = 0x0;
    return true;
  }

  function addAction(bytes32 name, address addr) returns (bool) {
    address actionDb = ContractProvider(DOUG).contracts("actiondb");
    if (actionDb == 0x0){
      return false;
    }
    bool res = ActionDb(actionDb).addAction(name,addr);
    return res;
  }

  function removeAction(bytes32 name) returns (bool) {
    address actionDb = ContractProvider(DOUG).contracts("actiondb");
    if (actionDb == 0x0){
      return false;
    }
    bool res = ActionDb(actionDb).removeAction(name);
    return res;
  }

  // Validate can be called by a contract like the bank to check if the
  // contract calling it has permissions to do so.
  function validate(address addr) constant returns (bool) {
    return addr == activeAction;
  }

}
```

Here is the new bank:

```javascript
// Interaction with the action manager.
contract Validator {
  function validate(address addr) constant returns (bool) {}
}

// The Bank contract - now inherits DougEnabled
contract Bank is DougEnabled {

  mapping(address => uint) public balance;

  // Endow an address with coins.
  function endow(address addr, uint amount) returns (bool) {
    address actns = ContractProvider(DOUG).contracts("actions");
    if (actns == 0x0){
      return false;
    }

    Validator v = Validator(actns);
    // If the sender is not validated successfully, break.
    if (!v.validate(msg.sender)){
      return false;
    }
    balance[addr] += amount;
    return true;
  }

  // Charge an account 'amount' number of coins.
  function charge(address addr, uint amount) returns (bool){
    address actns = ContractProvider(DOUG).contracts("actions");
    if (actns == 0x0){
      return;
    }

    Validator v = Validator(actns);
    // If the sender is not validated successfully, break.
    if (!v.validate(msg.sender)){
      return false;
    }

    if (balance[addr] < amount){
      return false;
    }

    balance[addr] -= amount;
    return true;
  }

}
```

What we have now is a system that allows us to add contracts (any contracts) to DOUG, and actions. The contracts can not be called except through actions, which means that we can control who gets to call the contracts by controlling who gets to execute actions, and since all actions are run in the same way it will be easy.

There is other benefits to a system like this as well, for example PRODOUG used the fact that all transactions went through the action manager to log them. The log included data such as the caller address, which action was called, the number of the block in which the tx was added, etc. This is good if you want to keep track of what's going on.

## Locking Things Down

The last thing we have to fix is access to DOUG and the action manager. It is true that the bank and other contracts must be called via actions, but anyone is allowed to add and remove actions, and also to add and remove contracts from DOUG. We're going to start by adding a simple permissions contract that we can use to set permissions for accounts. It'll be registered with DOUG under the name "perms". We're then going to add functions to actions where permissions can be gotten and set. Finally we will complement the system with the following basic actions:

- add action
- remove action
- add contract
- remove contract
- set account permissions
- modify action permissions

Note that there will be an add action action. It will be added to the action database upon creation, but can be replaced later (through the add action action itself).

*Pro tip: Don't remove the add action action.*

```javascript
// Interaction with the action manager.
contract Validator {
  function validate(address addr) constant returns (bool) {}
}

// The Permissions contract
contract Permissions is DougEnabled {

  // This is where we keep all the permissions.
  mapping (address => uint8) public perms;

  function setPermission(address addr, uint8 perm) returns (bool) {
    address actns = ContractProvider(DOUG).contracts("actions");
    if (actns == 0x0){
      return false;
    }
    Validator v = Validator(actns);
    // If the sender is not validated successfully, break.
    if (!v.validate(msg.sender)) {
      return false;
    }

    perms[addr] = perm;
  }

}
```

Next we will modify the actions template so that it is possible to get and set the permissions required to execute them. We will add the following functions to the interface:

```javascript
function permission(address addr) constant returns (uint) {}
function setPermission(uint8 permVal) returns (bool) {}
```

This is how we'd update the action managers execute function.

```javascript
// For getting permissions.
contract Permissioner {
  function perms(address addr) constant returns (uint8) { }
}

function execute(bytes32 actionName, bytes data) returns (bool) {

  ...

  // Permissions stuff
  address pAddr = ContractProvider(DOUG).getContract("perms");
  // If no permissions contract is added, then no permissions are required.
  if(pAddr != 0x0){
    Permissioner p = Permissioner(pAddr);

    // First we check the permissions of the account that's trying to execute the action.
    uint8 perm = p.getPermission(msg.sender);
    // Now we check the permission that is required to execute the action.
    uint8 permReq = Action(actn).permission();
    // Very simple system.
    if (perm < permReq){
      return false;
    }
  }

  // Proceed to execute the action.

  ...

}
```

## The Doubly Linked List

Before moving on to assembling the final contracts, we need to address something important that we haven't touched upon yet. If we look at Doug, or the action database, or any database contract for that matter, what bad thing do they all have in common? Well, the fact that we have no way of getting a collection of all the entries in the mappings. We have to get entries (such as contracts in the case of Doug) by key. The `mapping` type that backs all these databases has no built in iterator or function to get all elements. One way of adding these features to a mapping it by wrapping it inside a linked list data-structure.

The doubly linked list over a `mapping` provides many benefits. We can add and remove elements dynamically. We can get elements by key. All of these operations are O(1) so it is cheap with regards to computation. The drawback is that it adds extra data to storage, which is not insignificant.

{{< image src="/images/docs/LinkedList.png" >}}

So, what do we need to add?

### Step 1

First we need to add three additional fields to the contract - the size of the list, and references to the current head and tail. Let us start with a "generic" linked list contract that uses addresses as keys, and a fixed-length string as the value.

```javascript
contract DoublyLinkedList {
  uint size;
  address tail;
  address head;
  mapping(address => bytes32) elements;
}
```

### Step 2

To keep references to the previous and next element, we need to switch out the bytes32 value with a struct, like this:

```javascript
contract DoublyLinkedList {

  struct Element {
    address previous;
    address next;

    bytes32 data;
  }

  uint size;
  address tail;
  address head;
  mapping(address => Element) elements;
}
```

### Step 3

Now we need to implement the logic for adding and removing elements. Let's start with add. We're going to add elements as the new `head`, and the adding logic for an element is easy: Either the list is **empty**, which means the new element becomes both tail and head, or it is **non empty** and it becomes the new head.

When it comes to linking, we have the same thing. Either the list is empty, and no linking takes place, or the list is non empty and we must do the following steps:

Add the new element as the **next** element of of the current head, and add the current head as the **previous** element of the new element.

Assuming we don't allow elements to be over-written, and we use the mapping as a regular mapping, this is what the contract could look like with an add element function:

```javascript
contract DoublyLinkedList {

  struct Element {
    address previous;
    address next;

    bytes32 data;
  }

  uint size;
  address tail;
  address head;
  mapping(address => Element) elements;

  function addElement(address key, bytes32 data) returns (bool){
    Element elem = elements[key];
    // Check that the key is not already taken. We have no null-check for structs atm., so
    // we need to check the fields inside the structs to verify. This works if the field we
    // check is not allowed to be the null value (which would be "" in the case of strings).
    if(elem.data != ""){
      return false;
    }

    elem.data = data;

      // Two cases - empty or not.
      if(size == 0){
        tail = key;
        head = key;
      } else {
        // Link
        elements[head].next = key;
        elem.previous = head;
        // Set this element as the new head.
        head = key;
      }
      // Regardless of case, increase the size of the list by one.
      size++;
       return true;
  }
}
```

All in all, this is not too much code to add, and it's fairly straight forward. When it comes to removal, it's a bit more complicated. We need to consider three basic cases.

Case 1 is that the element we're removing is the only element in the list. In this case we need to set both head and tail to the null value, set size to 0, and remove the element data itself.

Case 2 is that the element is the head. That means we only have to modify the head field, and only the **next** field of one element - namely the element that is the current ones **previous**.

Case 3 is that the element is the tail, in which case it's similar.

Finally, case 4 is if this element is neither head nor tail. In this case the head and tail fields will not be touched, but we need to link "around" this element by changing the **previous** of this ones **next**, and the **next** of this ones **previous**. Here is the final contract with an add and remove function. We also add a special accessor that only gets the `data`, and not the entire element.

```javascript
contract DoublyLinkedList {

  struct Element {
    address previous;
    address next;

    bytes32 data;
  }

  // Make these public.
  uint public size;
  address public tail;
  address public head;
  mapping(address => Element) elements;

  function getData(address key) returns (bytes32){
    return elements[key].data;
  }

  function getElement(address key) constant returns (Element){
    return elements[key];
  }

  function addElement(address key, bytes32 data) returns (bool){
    Element elem = elements[key];
    // Check that the key is not already taken. We have no null-check for structs atm., so
    // we need to check the fields inside the structs to verify. This works if the field we
    // check is not allowed to be null (which would be 0 or 0x0 in the case of addresses).
    if(elem.data != ""){
      return false;
    }

    elem.data = data;

      // Two cases - empty or not.
      if(size == 0){
        tail = key;
        head = key;
      } else {
        // Link
        elements[head].next = key;
        elem.previous = head;
        // Set this element as the new head.
        head = key;
      }
      // Regardless of case, increase the size of the list by one.
      size++;
       return true;
  }


    function removeElement(address key) returns (bool result) {

       Element elem = elements[key];

      // If no element - return false. Nothing to remove.
      if(elem.data == ""){
        return false;
      }

    // If this is the only element.
      if(size == 1){
        tail = 0x0;
        head = 0x0;
      // If this is the head.
      } else if (key == head){
        // Set this ones 'previous' to be the new head, then change its
        // next to be null (used to be this one).
        head = elem.previous;
        elements[head].next = 0x0;
      // If this one is the tail.
      } else if(key == tail){
        tail = elem.next;
        elements[tail].previous = 0x0;
      // Now it's a bit tougher. Getting here means the list has at least 3 elements,
      // and this element must have both a 'previous' and a 'next'.
      } else {
        address prevElem = elem.previous;
        address nextElem = elem.next;
        elements[prevElem].next = nextElem;
        elements[nextElem].previous = prevElem;
      }
      // Regardless of case, we will decrease the list size by 1, and delete the actual entry.
      size--;
      delete elements[key];
      return true;
    }

}
```

That is all we need for a basic implementation.

To read this list from javascript, a simple loop could look like this:

```javascript
function getAllElements(){
  var list = [];
  var tail = listContract.tail();
  // isZero should basically just check if the hex-string evaluates to 0. Personally
  // i use bignumber.js for this, and it is included in the string math library
  // node.js decerver will have.
  if(isZero(tail)){
    return list;
  }
  var currentKey = tail;
  while(!isZero(currentKey)){
    var elem = listContract.getElement(currentKey);
    list.push(elem);
    // Slot 1 is 'next', 0 is previous, etc.
    currentKey = elem[1];
  }
  return list;
}
```

Note that accessing the element data by index is a bit ugly. Personally I use the json ABI to generate objects of the returned data with the proper names etc., using something like this (which will also be in Monax.js):

```javascript
// fName = function name.
function JsonAdapterOut(abi, fName){
  var outputs;
  // Check abi until we find the outputs array for the given function.
  for(var i = 0; i < abi.length; i++){
    var func = abi[i];
    if(abi[i].name.indexOf(fName) > -1){
      outputs = abi[i].outputs;
      break;
    }
  }
  if(outputs === null){
    window.alert("Failed to register json adapter");
  }

  // Syntax would be 'var funcOutputObj = jsonAdapter.convert(theContract.fName(arg0,arg1,...));'
  this.convert = function(data){
    var ret = {};
    for(var i = 0; i < outputs.length; i++){
      ret[outputs[i].name] = data[i].toString();
    }
    return ret;
  }
};
```

A note on generic types: Linked lists can not be fully generic right now. It would be doable in theory, if the key and data field in the Element struct were both `bytes` objects, but keys must be elementary types right now. Also, it would be hard to work with and document lists of that kind. Using bytes might be the way of doing it though, until/unless generics is added to Solidity, which is probably far into the future. What this means is a linked list generally has to be tailored for the contract that extends it.

Finally, since linked lists adds to the complexity of a big set of new contracts they will not be added to the finished contracts; instead there is a regular linked-list Doug contract included in the finished contracts section that can be used as a model. In part 3 it will use only linked lists.

### Wrapping up

Before assembling a list of the final contracts, we need to do some final modifications.

Doug will have to be modified. We need it to validate the account when someone is trying to add a contract. This is a bit weird, because how then would you go about adding the action manager contract? One way is to check if and action manager has been added. If there is no action manager then just allow anything. Adding the action manager is what you do to lock the system down. Also, what about removing? How do we remove Doug? Whoever is allowed to do that can kill the entire system with one press of a button, so this would often have to be regulated somehow, but if it's a normal dapp that has an owner it could be as easy as giving the owner the exclusive right to kill the DOUG contract. It does not have to be the same in every system.

Keep in mind, this is just a basic action driven architecture. PRODOUG for example had voting. This ment actions could sometimes  not be carried out directly, instead the action would spawn a copy of itself and be kept in a temporary list until the vote was done. Those types of actions had an init function where all the parameters was set, and then an execute function that was carrried out when a vote was concluded. The way it worked with permissions was that actions did not return a number when asked for the required permission but a name of a poll type. These poll types was kept in a list in a different manager that handled polls. Sometimes the polls were automatic (based on some user property) and sometimes there was a full-on vote with time limits, a quorums and other things. In hindsight, I think it would have been better to allow those type of actions to just store the indata in an indexed list of some sort, to keep track of which data belonged to which caller, until the vote has been resolved. CREATE calls which are very expensive on gas-enabled chains so short lived objects (poltergeists) should generally be kept at a minimum.

Finally, this system is still a bit tainted by the low level system it came out of.

### The finished contracts

Gonna throw in a few actions for locking and unlocking of the actionmanager as well as some extra logging stuff. It's good to be able to do that.

**Pure interfaces**

```javascript
contract ContractProvider {
    function contracts(bytes32 name) returns (address){}
}

contract Permissioner {
    function perms(address addr) constant returns (uint8) { }
}

contract Validator {
  function validate(address addr) constant returns (bool) {}
}

contract Charger {
  function charge(address addr, uint amount) returns (bool) {}
}

contract Endower {
  function endow(address addr, uint amount) returns (bool) {}
}
```

## Base Contracts

```javascript
contract DougEnabled {
    address DOUG;

    function setDougAddress(address dougAddr) returns (bool result){
        // Once the doug address is set, don't allow it to be set again, except by the
        // doug contract itself.
        if(DOUG != 0x0 && dougAddr != DOUG){
            return false;
        }
        DOUG = dougAddr;
        return true;
    }

    // Makes it so that Doug is the only contract that may kill it.
    function remove(){
        if(msg.sender == DOUG){
            selfdestruct(DOUG);
        }
    }

}

contract ActionManagerEnabled is DougEnabled {
    // Makes it easier to check that action manager is the caller.
    function isActionManager() internal constant returns (bool) {
        if(DOUG != 0x0){
            address am = ContractProvider(DOUG).contracts("actions");
            if (msg.sender == am){
                return true;
            }
        }
        return false;
    }
}

contract Validee {
    // Makes it easier to check that action manager is the caller.
    function validate() internal constant returns (bool) {
        if(DOUG != 0x0){
            address am = ContractProvider(DOUG).contracts("actions");
            if(am == 0x0){
              return false;
            }
            return Validator(am).validate(msg.sender);
        }
        return false;
    }
}
```

### ActionDB

```javascript
contract ActionDb is ActionManagerEnabled {

    // This is where we keep all the actions.
    mapping (bytes32 => address) public actions;

    // To make sure we have an add action action, we need to auto generate
    // it as soon as we got the DOUG address.
    function setDougAddress(address dougAddr) returns (bool result) {
      super.setDougAddress(dougAddr);

      var addaction = new ActionAddAction();
      // If this fails, then something is wrong with the add action contract.
      // Will be events logging these things in later parts.
      if(!DougEnabled(addaction).setDougAddress(dougAddr)){
          return false;
      }
      actions["addaction"] = address(addaction);
    }

    function addAction(bytes32 name, address addr) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        // Remember we need to set the doug address for the action to be safe -
        // or someone could use a false doug to do damage to the system.
        // Normally the Doug contract does this, but actions are never added
        // to Doug - they're instead added to this lower-level CMC.
        bool sda = DougEnabled(addr).setDougAddress(DOUG);
        if(!sda){
          return false;
        }
        actions[name] = addr;
        return true;
    }

    function removeAction(bytes32 name) returns (bool) {
        if (actions[name] == 0x0){
            return false;
        }
        if(!isActionManager()){
            return false;
        }
        actions[name] = 0x0;
        return true;
    }

}
```

### ActionManager

```javascript
contract ActionManager is DougEnabled {

  struct ActionLogEntry {
    address caller;
    bytes32 action;
    uint blockNumber;
    bool success;
  }

  bool LOGGING = true;

  // This is where we keep the "active action".
  // TODO need to keep track of uses of (STOP) as that may cause activeAction
  // to remain set and opens up for abuse. (STOP) is used as a temporary array
  // out-of bounds exception for example (or is planned to), which means be
  // careful. Does it revert the tx entirely now, or does it come with some sort
  // of recovery mechanism? Otherwise it is still super dangerous and should never
  // ever be used. Ever.
  address activeAction;

  uint8 permToLock = 255; // Current max.
  bool locked;

  // Adding a logger here, and not in a separate contract. This is wrong.
  // Will replace with array once that's confirmed to work with structs etc.
  uint public nextEntry = 0;
  mapping(uint => ActionLogEntry) public logEntries;

  function ActionManager(){
    permToLock = 255;
  }

  function execute(bytes32 actionName, bytes data) returns (bool) {
    address actionDb = ContractProvider(DOUG).contracts("actiondb");
    if (actionDb == 0x0){
      _log(actionName,false);
      return false;
    }

    address actn = ActionDb(actionDb).actions(actionName);
    // If no action with the given name exists - cancel.
    if (actn == 0x0){
      _log(actionName,false);
      return false;
    }

      // Permissions stuff
    address pAddr = ContractProvider(DOUG).contracts("perms");
    // Only check permissions if there is a permissions contract.
    if(pAddr != 0x0){
      Permissions p = Permissions(pAddr);

      // First we check the permissions of the account that's trying to execute the action.
      uint8 perm = p.perms(msg.sender);

      // Now we check that the action manager isn't locked down. In that case, special
      // permissions is needed.
      if(locked && perm < permToLock){
        _log(actionName,false);
        return false;
      }

      // Now we check the permission that is required to execute the action.
      uint8 permReq = Action(actn).permission();

      // Very simple system.
      if (perm < permReq){
        _log(actionName,false);
          return false;
      }

    }

    // Set this as the currently active action.
    activeAction = actn;
    // TODO keep up with return values from generic calls.
    // Just assume it succeeds for now (important for logger).
    actn.call(data);
    // Now clear it.
    activeAction = 0x0;
    _log(actionName,true);
    return true;
  }

  function lock() returns (bool) {
    if(msg.sender != activeAction){
      return false;
    }
    if(locked){
      return false;
    }
    locked = true;
  }

  function unlock() returns (bool) {
    if(msg.sender != activeAction){
      return false;
    }
    if(!locked){
      return false;
    }
    locked = false;
  }

  // Validate can be called by a contract like the bank to check if the
  // contract calling it has permissions to do so.
  function validate(address addr) constant returns (bool) {
    return addr == activeAction;
  }

  function _log(bytes32 actionName, bool success) internal {
    // TODO check if this is really necessary in an internal function.
    if(msg.sender != address(this)){
      return;
    }
    ActionLogEntry le = logEntries[nextEntry++];
    le.caller = msg.sender;
    le.action = actionName;
    le.success = success;
    le.blockNumber = block.number;
  }

}
```

### Doug

```javascript
contract Doug {

    address owner;

    // This is where we keep all the contracts.
    mapping (bytes32 => address) public contracts;

    // Constructor
    function Doug(){
        owner = msg.sender;
    }

    // Add a new contract to Doug. This will overwrite an existing contract.
    function addContract(bytes32 name, address addr) returns (bool result) {
    // Only do validation if there is an actions contract.
    var am = contracts["actions"];
    if(am != 0x0 || contracts["actionsdb"] == 0x0){
      // Check that the account trying to add a contract is a registered action.
          bool val = Validator(am).validate(msg.sender);
          if(!val){
            return false;
      }
       }
       DougEnabled de = DougEnabled(addr);
       // Don't add the contract if this does not work.
    if(!de.setDougAddress(address(this))) {
      return false;
    }
    contracts[name] = addr;
       return true;
  }

    // Remove a contract from Doug. We could also selfdestruct if we want to.
    function removeContract(bytes32 name) returns (bool result) {
       address cName = contracts[name];
       if (cName == 0x0){
           return false;
       }
       // Only do validation if there is an actions contract.
       var am = contracts["actions"];
    if(am != 0x0 || contracts["actionsdb"] == 0x0){
          // Check that the account trying to add a contract is a registered action.
          bool val = Validator(am).validate(msg.sender);
          if(!val){
            return false;
          }
        }
        // Kill any contracts we remove, for now.
        DougEnabled(cName).remove();
        contracts[name] = 0x0;
        return true;
    }

    function remove(){
        if(msg.sender == owner){
            selfdestruct(owner);
        }
    }

}
```

### Bank

```javascript
contract Bank is Validee {

  mapping(address => uint) balance;

  // Endow an address with coins.
  function endow(address addr, uint amount) returns (bool) {
    if (!validate()){
      return false;
    }
    balance[addr] += amount;
    return true;
  }

  // Charge an account 'amount' number of coins.
  function charge(address addr, uint amount) returns (bool){
    if (balance[addr] < amount){
      return false;
    }
    if (!validate()){
      return false;
    }
    balance[addr] -= amount;
    return true;
  }

}
```

### Permissions

```javascript
// The Permissions contract
contract Permissions is Validee {

    // This is where we keep all the permissions.
    mapping (address => uint8) public perms;

    function setPermission(address addr, uint8 perm) returns (bool) {
    if (!validate()){
      return false;
    }
    perms[addr] = perm;
    }

}
```

### Actions

```javascript
contract Action is ActionManagerEnabled, Validee {
  // Note auto accessor.
  uint8 public permission;

  function setPermission(uint8 permVal) returns (bool) {
    if(!validate()){
      return false;
    }
    permission = permVal;
  }
}

// Add action. NOTE: Overwrites currently added actions with the same name.
contract ActionAddAction is Action {

    function execute(bytes32 name, address addr) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address adb = dg.contracts("actiondb");
        if(adb == 0x0){
            return false;
        }
        return ActionDb(adb).addAction(name, addr);
    }

}

// Remove action. Does not allow 'addaction' to be removed, though that it can still
// be done by overwriting this action with one that allows it.
contract ActionRemoveAction is Action {

    function execute(bytes32 name) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address adb = dg.contracts("actiondb");
        if(adb == 0x0){
            return false;
        }
        if(name == "addaction"){
          return false;
        }
        return ActionDb(adb).removeAction(name);
    }

}

// Lock actions. Makes it impossible to run actions for everyone but the owner.
// It is good to unlock the actions manager while replacing parts of the system
// for example.
contract ActionLockActions is Action {

    function execute() returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address am = dg.contracts("actions");
        if(am == 0x0){
            return false;
        }
        return ActionManager(am).lock();
    }

}

// Unlock actions. Makes it possible for everyone to run actions.
contract ActionUnlockActions is Action {

    function execute() returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address am = dg.contracts("actions");
        if(am == 0x0){
            return false;
        }
        return ActionManager(am).unlock();
    }

}

// Add contract.
contract ActionAddContract is Action {

    function execute(bytes32 name, address addr) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        Doug d = Doug(DOUG);
        return d.addContract(name,addr);
    }

}

// Remove contract.
contract ActionRemoveContract is Action {

    function execute(bytes32 name) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        Doug d = Doug(DOUG);
        return d.removeContract(name);
    }

}

// The charge action.
contract ActionCharge is Action {

    function execute(address addr, uint amount) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address charger = dg.contracts("bank");
        if(charger == 0x0){
            return false;
        }
        return Charger(charger).charge(addr,amount);
    }

}

// The endow action.
contract ActionEndow is Action {

    function execute(address addr, uint amount) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address endower = dg.contracts("bank");
        if(endower == 0x0){
            return false;
        }
        return Endower(endower).endow(addr,amount);
    }

}

// The set user permission action.
contract ActionSetUserPermission is Action {

    function execute(address addr, uint8 perm) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address perms = dg.contracts("perms");
        if(perms == 0x0){
            return false;
        }
        return Permissions(perms).setPermission(addr,perm);
    }

}

// The set action permission. This is the permission level required to run the action.
contract ActionSetActionPermission is Action {

    function execute(bytes32 name, uint8 perm) returns (bool) {
        if(!isActionManager()){
            return false;
        }
        ContractProvider dg = ContractProvider(DOUG);
        address adb = dg.contracts("actiondb");
        if(adb == 0x0){
            return false;
        }
        var action = ActionDb(adb).actions(name);
        Action(action).setPermission(perm);
    }

}
```

Linked list Doug

```javascript
contract DougEnabled {
    function setDougAddress(address dougAddr) returns (bool result){}
    function remove(){}
}

//The Doug database contract.
contract DougDb {

     // List element
  struct Element {
    bytes32 prev;
    bytes32 next;
    // Data
    bytes32 contractName;
    address contractAddress;
  }

  uint public size;
  bytes32 public tail;
  bytes32 public head;
    mapping (bytes32 => Element) list;

  // Add a new contract. This will overwrite an existing contract. 'internal' modifier means
  // it has to be called by an implementing class.
  function _addElement(bytes32 name, address addr) internal returns (bool result) {
       Element elem = list[name];

      elem.contractName = name;
      elem.contractAddress = addr;

      // Two cases - empty or not.
      if(size == 0){
        tail = name;
        head = name;
      } else {
        list[head].next = name;
        list[name].prev = head;
        head = name;
      }
      size++;
       return true;
    }

    // Remove a contract from Doug (we could also selfdestruct the contract if we want to).
    function _removeElement(bytes32 name) internal returns (bool result) {

       Element elem = list[name];
      if(elem.contractName == ""){
        return false;
      }

      if(size == 1){
        tail = "";
        head = "";
      } else if (name == head){
        head = elem.prev;
        list[head].next = "";
      } else if(name == tail){
        tail = elem.next;
        list[tail].prev = "";
      } else {
        bytes32 prevElem = elem.prev;
        bytes32 nextElem = elem.next;
        list[prevElem].next = nextElem;
        list[nextElem].prev = prevElem;
      }
      size--;
      delete list[name];
      return true;
  }

  // Should be safe to update to returning 'Element' instead
  function getElement(bytes32 name) constant returns (bytes32 prev, bytes32 next, bytes32 contractName, address contractAddress) {

      Element elem = list[name];
      if(elem.contractName == ""){
        return;
      }
      prev = elem.prev;
      next = elem.next;
      contractName = elem.contractName;
      contractAddress = elem.contractAddress;
  }

}


/// @title DOUG
/// @author Andreas Olofsson
/// @notice This contract is used to register other contracts by name.
/// @dev Stores the contracts as entries in a doubly linked list, so that
/// the list of elements can be gotten.
contract Doug is DougDb {

  address owner;

     // When adding a contract.
  event AddContract(address indexed caller, bytes32 indexed name, uint16 indexed code);
  // When removing a contract.
  event RemoveContract(address indexed caller, bytes32 indexed name, uint16 indexed code);

    // Constructor
    function Doug(){
        owner = msg.sender;
    }

    /// @notice Add a contract to Doug. This contract should extend DougEnabled, because
    /// Doug will attempt to call 'setDougAddress' on that contract before allowing it
    /// to register. It will also ensure that the contract cannot be selfdestructed by anyone
    /// other than Doug. Finally, Doug allows over-writing of previous contracts with
    /// the same name, thus you may replace contracts with new ones.
    /// @param name The bytes32 name of the contract.
    /// @param addr The address to the actual contract.
    /// @returns boolean showing if the adding succeeded or failed.
    function addContract(bytes32 name, address addr) returns (bool result) {
       // Only the owner may add, and the contract has to be DougEnabled and
       // return true when setting the Doug address.
    if(msg.sender != owner || !DougEnabled(addr).setDougAddress(address(this))){
      // Access denied. Should divide these up into two maybe.
      AddContract(msg.sender, name, 403);
      return false;
    }
       // Add to contract.
       bool ae = _addElement(name, addr);
       if (ae) {
          AddContract(msg.sender, name, 201);
       } else {
          // Can't overwrite.
          AddContract(msg.sender, name, 409);
       }
       return ae;
  }

    /// @notice Remove a contract from doug.
    /// @param name The bytes32 name of the contract.
    /// @returns boolean showing if the removal succeeded or failed.
    function removeContract(bytes32 name) returns (bool result) {
        if(msg.sender != owner){
            RemoveContract(msg.sender, name, 403);
            return false;
        }
        bool re = _removeElement(name);
        if(re){
          RemoveContract(msg.sender, name, 200);
        } else {
          // Can't remove, it's already gone.
          RemoveContract(msg.sender, name, 410);
        }
        return re;
    }

    /// @notice Gets a contract from Doug.
    /// @param name The bytes32 name of the contract.
    /// @returns The address of the contract. If no contract with that name exists, it will
    /// return zero.
    function contracts(bytes32 name) returns (address addr){
      return list[name].contractAddress;
    }

    /// @notice Remove (selfdestruct) Doug.
    function remove(){
        if(msg.sender == owner){
            // Finally, remove doug. Doug will now have all the funds of the other contracts,
            // and when suiciding it will all go to the owner.
            selfdestruct(owner);
        }
    }

}
```


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Solidity Tutorials](/docs/solidity/)
