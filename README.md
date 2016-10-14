# duet
SE 464 project

## Set up

Start off by setting up your GOPATH and updating your PATH

```
export GOPATH=$HOME/go # Go packages will be installed here from github
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin # Built Go files will go here
```

## Deploy

SSH into the server, and then go to `/var/www/duet/`.

Run
```
sudo service nginx restart
godep go run main.go
```

## Updating Dependencies
If new packages are installed, run
```
godep save
```
This saves the exact version of the dependency used.
