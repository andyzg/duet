# duet
SE 464 project

## Set up

Start off by setting up your `GOPATH` and updating your `PATH`

```
export GOPATH=$HOME/go # Go packages will be installed here from github
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin # Built Go files will go here
```

Also install `godep` by running `go get github.com/tools/godep`.

## Deploy
Make sure this repository is in your `GOPATH` then run
```
godep get
godep go build
./duet &
```

This serves the API on port 8080. graphiql, a GraphQL explorer, is located at `:8080/` and the GraphQL endpoint
is `:8080/graphql`.

## Updating Dependencies
If new packages are installed, run `godep save`. This saves the exact version of the dependency used.

To fetch dependencies run `godep get`.
