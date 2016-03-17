# gosass [![Build Status](https://travis-ci.org/dailymuse/gosass.png)](https://travis-ci.org/dailymuse/gosass) #

## About ##

Gosass is a fast sass compiler. We made this because the normal ruby compiler
ran too slowly, and the native compiler (`sassc`) didn't have all the features
we needed (directory compilation and watching.)

## Installation ##

Gosass depends on `sassc`, so
[install that first](https://github.com/sass/sassc). On mac, you can use
homebrew: `brew install sassc`.

Then install `gosass`:

```bash
go get github.com/dailymuse/gosass
pushd $GOPATH/src/github.com/dailymuse/gosass
make install
popd
```

## Example Usage ##

```bash
# Compile an individual file
gosass -input file.scss -output file.css

# Compile a directory
gosass -input sass/ -output css/

# Live-compile a directory
gosass -input sass/ -output css/ -watch
```
