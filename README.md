# duet
SE 464 project

## Set up

Start off by setting up your GOPATH and updating your PATH

```
export GOPATH=$HOME/go # Go packages will be installed here from github
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin # Built Go files will go here
```

## Docker
With docker installed, do
```
cd go
docker build -t duet-go .
docker run -p 8080:8080 duet-go
```

nginx should be configured to serve port 8080 to api.helloduet.com.
