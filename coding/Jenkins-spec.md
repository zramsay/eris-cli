Repos that are built by jenkins have three requirements:

1. The name should be in the form of {eris,monax}-<binary_name> example: github.com/eris-ltd/eris-pm would generate a binary called pm.
2. use glide, so we can consistently manipulate dependencies as needed.
3. the dockerfile which builds the binary should place the binary in the container via this sample dockerfile code:

> ENV INSTALL_BASE /usr/local/bin  
> ...  
> ENV REPO $GOPATH/src/github.com/eris-ltd/eris-pm  
> COPY . $REPO  
> ...  
> WORKDIR $REPO/cmd/epm  
> RUN go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/epm  
> RUN chown --recursive $USER:$USER $REPO  
