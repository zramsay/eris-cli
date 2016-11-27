package initialize

// TODO [zr] use templates

func defaultDeveloperAccountType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "Developer"

definition = """
Users who have a key which is registered with developer privileges can send
tokens; call contracts; create contracts; create accounts; use the name registry;
and modify account's roles.
"""

typical_user = """
Generally the development team seeking to build the application on top of the
given chain would be within the group. If this is a multi organizational
chain then developers from each of the stakeholders should generally be registered
within this group, although that design is up to you.
"""

default_number = 6

default_tokens = 9999999999

default_bond = 0

[perms]
root = 0
send = 1
call = 1
create_contract = 1
create_account = 1
bond = 0
name = 1
has_base = 0
set_base = 0
unset_base = 0
set_global = 0
has_role = 1
add_role = 1
rm_role = 1
`
}

func defaultFullAccountType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "Full"

definition = """
Users who have a key which is registered with root privileges can do everything
on the chain. They have all of the permissions possible. These users are also
bonded at the genesis time, so these should be used only for simple chains with
a few nodes who will be on during the prototyping session.
"""

typical_user = """
If you are making a small chain just to play around then usually you would
give all of the accounts needed for your experiment full accounts.

If you are making a more complex chain, don't use this account type.
"""

default_number = 1

default_tokens = 99999999999999

default_bond = 9999999999

[perms]
root = 1
send = 1
call = 1
create_contract = 1
create_account = 1
bond = 1
name = 1
has_base = 1
set_base = 1
unset_base = 1
set_global = 1
has_role = 1
add_role = 1
rm_role = 1
`
}

func defaultParticipantAccountType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "Participant"

definition = """
Users who have a key which is registered with participant privileges can send
tokens; call contracts; and use the name registry
"""

typical_user = """
Generally the number of participants in your chain who do not need elevated
privileges should be given these keys.

Usually this group will have the most number of keys of all of the groups.
"""

default_number = 25

default_tokens = 9999999999

default_bond = 0

[perms]
root = 0
send = 1
call = 1
create_contract = 0
create_account = 0
bond = 0
name = 1
has_base = 0
set_base = 0
unset_base = 0
set_global = 0
has_role = 1
add_role = 0
rm_role = 0
`
}

func defaultRootAccountType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "Root"

definition = """
Users who have a key which is registered with root privileges can do everything
on the chain. They have all of the permissions possible.
"""

typical_user = """
If you are making a small chain just to play around then usually you would
give all of the accounts needed for your experiment root privileges (unless you
were testing different) privilege types.

If you are making a more complex chain, then you would usually have a few
keys which have registered root permissions and as such will act in a capacity
similar to a network administrator in other data management situations.
"""

default_number = 3

default_tokens = 9999999999

default_bond = 0

[perms]
root = 1
send = 1
call = 1
create_contract = 1
create_account = 1
bond = 1
name = 1
has_base = 1
set_base = 1
unset_base = 1
set_global = 1
has_role = 1
add_role = 1
rm_role = 1
`
}

func defaultValidatorAccountType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "Validator"

definition = """
Users who have a key which is registered with validator privileges can
only post a bond and begin validating the chain. This is the only privilege
this account group gets.
"""

typical_user = """
Generally the marmots recommend that you put your validator nodes onto a cloud
(IaaS) provider so that they will be always running.

We also recommend that if you are in a multi organizational chain then you would
have some separation of the validators to be ran by the different organizations
in the system.
"""

default_number = 7

default_tokens = 9999999999

default_bond = 9999999998

[perms]
root = 0
send = 0
call = 0
create_contract = 0
create_account = 0
bond = 1
name = 0
has_base = 0
set_base = 0
unset_base = 0
set_global = 0
has_role = 0
add_role = 0
rm_role = 0
`
}
