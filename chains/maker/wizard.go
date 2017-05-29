package maker

import (
	"fmt"

	"github.com/monax/monax/definitions"
)

func ChainsMakeWelcome() string {
	return `Welcome! I'm the marmot that helps you make your chain.

monax chains is your gateway to permissioned, smart contract compatible chains.
There is a bit of complexity around getting these chains set up. It is my
marmot-y task to make this as easy as we can.

First we will run through the monax chains typical account types and I'll ask
you how many of each account type you would like.

After that you will have an opportunity to create your own account types and
add those groupings into the genesis.json.

Remember, I'm only useful when you are making a new chain. After your chain
is established then you can modify the permissions using other monax tooling.

Are you ready to make your own chain (y/n)? `
}

func ChainsMakePrelimQuestions() map[string]string {
	questions := map[string]string{
		"dryrun": "Would you like to review each of the account groups before making the chain? (y/n) ",
		"tokens": "Do you care about setting the number of tokens for each account type? If not, the marmots will set reasonable defaults for you. (y/n) ",
		"manual": "After the built in account groups would you like to make your manual (own) account groups with fine tuned permissions? (y/n) ",
	}
	return questions
}

func AccountTypeIntro(account *definitions.MonaxDBAccountType) string {
	return fmt.Sprintf(`
The %s Group.

Group Definition:

%s

Who Should Get These?:

%s

How many keys do you want in the %s Group? (%d) `, account.Name, account.Description, account.TypicalUser, account.Name, account.DefaultNumber)
}

func AccountTypeTokens(account *definitions.MonaxDBAccountType) string {
	return fmt.Sprintf("How many tokens should each key in the %s Group be given? ", account.Name)
}

func AccountTypeToBond(account *definitions.MonaxDBAccountType) string {
	return fmt.Sprintf("This group appears to be a validating group. How many tokens would you like the %s Group to bond?", account.Name)
}

func AccountTypeManualIntro() string {
	return `Make Your Own Group.

Group Definition:

You will next be asked a series of questions regarding this group as to
what permissions you want.

Who Should Get These?:

Don't ask us, you are the one that wanted "manual" :-)

For more on monax chains permissions see here:

https://monax.io/platform/db

or

https://github.com/monax/burrow

How many keys do you want in *this* manual group? (You can make more than one manual group)`
}

func AccountTypeManualTokens() string {
	return "How many tokens should each key in this manual group be given? "
}

func AccountTypeManualToBond() string {
	return "How many tokens should each key in this manual group bond (0 means it shouldn't post bond)? "
}

// todo: this should autopopulate, but for later.
func AccountTypeManualPerms() []string {
	return []string{
		"root",
		"send",
		"call",
		"createContract",
		"createAccount",
		"bond",
		"name",
		"hasBase",
		"setBase",
		"unsetBase",
		"setGlobal",
		"hasRole",
		"addRole",
		"removeRole",
	}
}

func AccountTypeManualPermsQuestion(perm string) string {
	return fmt.Sprintf("Does the group have %s privileges? 0 means no/false (and is the default); 1 means yes/true (0/1) ", perm)
}

func AccountTypeManualSave() string {
	return "If you would like to save this account type for future use, what name should we give it? (leave blank and the marmots won't save it) "
}

func AccountTypeManualAnother() string {
	return "Do you want to make another manual group (y/n) "
}
