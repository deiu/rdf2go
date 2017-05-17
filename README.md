# rdf2go

[![Build Status](https://api.travis-ci.org/deiu/rdf2go.svg?branch=master)](https://travis-ci.org/deiu/rdf2go)
[![Coverage Status](https://coveralls.io/repos/github/deiu/rdf2go/badge.svg?branch=master)](https://coveralls.io/github/deiu/rdf2go?branch=master)

Native golang parser/serializer from/to Turtle and JSON-LD.

# Installation

Just go get it!

`go get -u github.com/deiu/rdf2go`

# Example usage

## Working with graphs

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g, err := NewGraph(baseUri)
if err != nil {
	// deal with err
}

// Add a few triples to the graph
triple1 := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
g.Add(triple1)
triple2 := NewTriple(NewResource("a"), NewResource("d"), NewResource("e"))
g.Add(triple2)

// Get length of Graph (nr of triples)
g.Len() // -> 2

// Dump graph contents to NTriples
out := g.String()
// <a> <b> <c> .
// <a> <d> <e> .

// Delete a triple
g.Remove(triple2)
```

## Looking up triples from the graph

### Returning a single match

The `g.One()` method returns the first triple that matches against any (or all) of Subject, Predicate, Object patterns.

```golang
// Create a new graph
g, _ := NewGraph("https://example.org")

// Add a few triples
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("c")))

// Look up one triple matching the given subject
triple := g.One(NewResource("a"), nil, nil) // -> <a> <b> <c> .

// Look up one triple matching the given predicate
triple = g.One(nil, NewResource("b"), nil) // -> <a> <b> <c> .

// Look up one triple matching the given object
triple = g.One(nil, nil, NewResource("c")) // -> <a> <b> <c> .

// Look up one triple matching the given subject and predicate
triple = g.One(NewResource("a"), NewResource("b"), nil) // -> <a> <b> <c> .

// Look up one triple matching the a bad predicate
triple = g.One(nil, NewResource("z"), nil) // -> nil
```

### Returning a list of matches

Similar to `g.One()`, `g.All()` returns all triples that match the given pattern.

```golang
// Create a new graph
g, _ := NewGraph("https://example.org")

// Add a few triples
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("c")))
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("d")))

// Look up one triple matching the given subject
triples := g.All(nil, nil, NewResource("c")) //
for triple := range triples {
	triple.String()
}
// Returns a single triple that matches object <c>:
// <a> <b> <c> .

triples = g.All(nil, NewResource("b"), nil)
for triple := range triples {
	triple.String()
}
// Returns all triples that match subject <b>: 
// <a> <b> <c> .
// <a> <b> <d> .
```

## Different types of terms (resources)

### IRIs

```golang
// Create a new IRI
iri := NewResource("https://example.org")
iri.String() // -> <https://example.org>
```

### Literals

```golang
// Create a new simple Literal
lit := NewLiteral("hello world")
lit.String() // -> "hello word"

// Create a new Literal with language tag
lit := NewLiteralWithLanguage("hello world", "en")
lit.String() // -> "hello word"@en

// Create a new Literal with a data type
lit := NewLiteralWithDatatype("newTypeVal", NewResource("https://datatype.com"))
lit.String() // -> "newTypeVal"^^<https://datatype.com>
```

### Blank Nodes

```golang
// Create a new Blank Node with a given ID
bn := NewBlankNode(9)
bn.String() // -> "_:n9"

// Create an anonymous Blank Node with a random ID
abn := NewAnonNode()
abn.String() // -> "_:n192853"
```


## Parsing data

The parser takes an `io.Reader` as first parameter, and the string containing the mime type as the second parameter.

Currently, the supported parsing formats are Turtle (with mime type `text/turtle`) and JSON-LD (with mime type `application/ld+json`).

### Parsing Turtle from an io.Reader

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g, _ := NewGraph(baseUri)

// r is of type io.Reader
g.Parse(r, "text/turtle")
```

### Parsing JSON-LD from an io.Reader

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g, _ := NewGraph(baseUri)

// r is an io.Reader
g.Parse(r, "application/ld+json")
```

### Parsing either Turtle or JSON-LD from a URI on the Web

In this case you don't have to specify the mime type, as the internal http client will try to content negotiate to either Turtle or JSON-LD. An error will be returned if it fails.

**Note:** The `NewGraph()` function accepts an optional parameter called `skipVerify` that is used to tell the internal http client whether or not to ignore bad/self-signed server side certificates. By default, it will not check if you omit this parameter, or if you set it to `true`.

```golang
// Set a base URI
uri := "https://example.org/foo"

// Check remote server certificate to see if it's valid 
// (don't skip verification)
skipVerify := false

// Create a new graph. You can also omit the skipVerify parameter
// and accept invalid certificates (e.g. self-signed)
g, _ := NewGraph(uri, skipVerify)

err := g.LoadURI(uri)
if err != nil {
	// deal with the error
}
```


## Serializing data


The serializer takes an `io.Writer` as first parameter, and the string containing the mime type as the second parameter.

Currently, the supported serialization formats are Turtle (with mime type `text/turtle`) and JSON-LD (with mime type `application/ld+json`).


### Serializing to Turtle

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g, _ := NewGraph(baseUri)

triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
g.Add(triple)

// w is of type io.Writer
g.Serialize(w, "text/turtle")
```

### Serializing to JSON-LD

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g, _ := NewGraph(baseUri)

triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
g.Add(triple)

// w is of type io.Writer
g.Serialize(w, "application/ld+json")
```
