Pangolin
========

Pangolin is a database written in Go.

We have hopes of making this distributed some day.

## Roadmap
- [ ] Create a data directory
- [ ] Create a bucket
- [ ] Get buckets
- [ ] Upsert item
- [ ] Get item
- [ ] Delete bucket
- [ ] Listen on port 
- [ ] Serve a simple GUI
- [ ] CLI for running the database

## API 
GET `/` will return an overview of the database 

An example response would look like

```JSON
  {"version":"0.0.1","buckets":["none"]}
```

## Testing 
Install dependencies, then you can run 

`go test github.com/dylanlott/pangolin{,/routes}`

This will run all tests.

## About

Pangolin is a project written for an internal hackathon at Storj Labs.

### Indexing
Indexing is accomplished using a left-leaning red-black tree.
This achieves log2N searches and 2 log2N tree height.

- [https://en.wikipedia.org/wiki/Left-leaning_red%E2%80%93black_tree]()
- [http://www.cs.princeton.edu/~rs/talks/LLRB/LLRB.pdf]()

## License 
Pangolin is AGPL 3.0 Licensed
