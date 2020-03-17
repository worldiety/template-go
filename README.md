# template-go
A sample go project with our required minimal fuss.

## usage

There is a little setup tool, which assists you for generating the correct project setup:

```bash
export GOPATH=$(go env var GOPATH | tr -d '\n') && go get github.com/worldiety/template-go && $GOPATH/bin/template-go
```
