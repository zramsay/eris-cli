# Actions Specification

Display and Manage actions for various components of the Eris platform and for the platform itself.

Actions are bundles of commands which rely upon a project which is currently in scope or a any set of installed services. Actions are held in yaml, toml, or json action-definition files within the action folder in the eris tree (globally scoped actions) or in a directory pointed to by the actions field of the currently checked out project (project scoped actions). Actions are a sequence of commands which operate in a similar fashion to how a circle.yml file may operate.

# Managing Actions

## get

Retrieve an action file from the internet (utilizes git clone or ipfs). NOTE: This functionality is currently limited to github.com and IPFS.

## new

Create a new action definition file optionally from a template.

## add

Actions must be an array of executable commands which will be called according to the machine definition included in eris config. Actions may be stored in JSON, TOML, or YAML. Globally accessible actions are stored in the actions directory of the eris tree. Project accessible actions are stored in a directory pointed to by the actions field of the currently checked out project.

## ls

List all registered action definition files.

## do

Perform an action according to the action definition file.  (limit by machine definition specification)

## edit

Edit an action definition file in the default editor.

## rename

Rename an action.

## remove

Remove an action definition file.

# Action Definition Files

Action definitions are kept in JSON, TOML, or YAML formatting. All files which are validly formatted and within the `~/.eris/actions` directory will be made available globally across all projects registered with `eris`. In addition, [project definition files](project_definition_files) may contain an `actions` field which should point to a valid directory (usually the `./actions` folder of the application) containing action definition files. When a project has been `checked out` then the actions for that project are **brought into scope** and may be used by `eris`.

`eris` will read four fields from action definition files. More fields may be present, and none of the read fields are required.

## Fields

`maintainer` -- Maintainer **must** contain a string value. Usually this will be an identifier of the maintainer of the action (usually name and email).

`uri` -- URI **must** contain a string value. The universal resource indicator where others may find this action definition file (this may be an IPFS entry, a field in `etcb` or a repository online).

`name` -- Name **must** contain a string value. The action's name which will be recommended upon installation.

`action` -- Action **must** contain an array value. The action field contains the sequence of steps which are to be performed in order to facilitate (`do`) the desired action.

All steps will be performed within the same subshell which will be started with the global environment variables mandated in the `eris config` and, if a project is checked out, the project environment variables mandated in the project definition file. Although action steps are executed within a subshell, they will be ran by the same user and group which calls the parent `eris actions do NAME` command.

Actions are structured to be an efficient way to establish seqences of plumbing commands which are often strung together to make higher level functionality. As such **no special security** considerations are taken into account by the `eris actions` functionality.