# Monax Compilers

## How to contribute a language (for those who know little to no golang)

We marmots are big time tinkerers and always appreciate new toys to play with. The following is a means for those interested in contributing to a thousand chain world to best contribute their design for smart contract compilation. If your chain is not EVM compliant there may be a few more steps, be sure to contact the marmots in the [marmot den](https://slack.monax.io/).

### Step 1: Dockerize your smart contract language

The marmots are heavy users of containers and require a nice docker image to work off of. Steps for how to create a docker image can be found [here](https://docs.docker.com/engine/getstarted/).

### Step 2: Create your contract template

Our templating design is very simple to understand how we handle compilers. Simply add the name of your compiler plus Return into the Return struct like below.

```golang
// note that this can be found in template.go
package compilers

// An interface to denote a compiler. It implements one function, Compile, which takes in a slice of files and a version string
// and returns a Compile Return or an error
type Compiler interface {
	Compile(files []string, version string) (Return, error)
}

//Practicing inheritance, this struct gives us access to all types of returns
//This is written to be extendable to other compilers
type Return struct {
	SolcReturn
	//Enter your return struct here...
}

```

Now things get a little bit more interesting. Create a separate file `myLang.go`. From there create a struct for your returns just like you did with the above with all of your outputs that you might need for smart contract interaction. Note that this does get easier if your language has a json output format: 

```golang
package compilers

type MyLangReturn struct {
	Binaries string `json:bin`
	Abi      string `json:abi`
}

```

## Step 3: Design CLI interaction

The next step is going to be describing how your language's CLI will handle the inputs and ultimately sending the command off. It can be as simple as:

```golang
// this example includes json encoding so we're going to use this standard package to get a json return back into the monax CLI
import (
	"encoding/json"
)

type MyLangTemplate struct {
	// This is normally an area for how the job runner would handle the compiler, we're going to ignore it for now and touch on it 
	// in a separate document in the jobs readme
}

func (m *MyLangTemplate) Compile(files []string, version string) (Return, error) {
	// denote the name of your language and give it an image for your docker hub
	var myLang string = "myLang"
	var image string = "myDockerHub/myLang:" + version
	// initialize an empty return struct
	myReturn := &MyLangReturn{}
	// we're going to keep it simple and clean by creating an array of string with myLang at the front
	exec := []string{myLang}
	// add files to the end of the command
	exec = append(exec, files...)
	// and execute the command via our helper function for executing commands 
	// and get the output in a byte string 
	output, err := ExecuteCompilerCommand(image, exec)
	// handle any errors
	if err != nil {
		return Return{}, err
	}
	// unmarshall your output to json, handling errors along the way
	if err = json.Unmarshal(output, myReturn); err != nil {
		return Return{}, err
	}
	// and finally return!
	return Return{*myReturn}, nil
}
```

## Step 4: Head over to the Job runner 

Head over to the Job runner documentation and edit or make a job just for your compiler. Happy dev-ing! 