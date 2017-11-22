## distill
This is a transformer for converting BGPDump data to JSON.  Not very useful in it's current state
and I'm using it to learn some golang.

## How to test
`make test`

## How to run it
```sh
make build
./distill testdata/bview.20161001.0000.gz output.json
less output.json
```
