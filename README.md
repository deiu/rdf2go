# rdf2go

[![CI](https://github.com/deiu/rdf2go/actions/workflows/ci.yml/badge.svg)](https://github.com/deiu/rdf2go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/deiu/rdf2go.svg)](https://pkg.go.dev/github.com/deiu/rdf2go)
[![Go Report Card](https://goreportcard.com/badge/github.com/deiu/rdf2go)](https://goreportcard.com/report/github.com/deiu/rdf2go)

Native golang parser/serializer from/to Turtle and JSON-LD.

# Installation

```
go get github.com/deiu/rdf2go
```

Requires Go 1.23 or newer.

# Example usage

## Working with graphs

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g := NewGraph(baseUri)

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
g := NewGraph("https://example.org")

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
g := NewGraph("https://example.org")

// Add a few triples
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("c")))
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("d")))

// Look up one triple matching the given subject
triples := g.All(nil, nil, NewResource("c")) //
for _, triple := range triples {
	triple.String()
}
// Returns a single triple that matches object <c>:
// <a> <b> <c> .

triples = g.All(nil, NewResource("b"), nil)
for _, triple := range triples {
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
g := NewGraph(baseUri)

// r is of type io.Reader
g.Parse(r, "text/turtle")
```

### Parsing JSON-LD from an io.Reader

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g := NewGraph(baseUri)

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
g := NewGraph(uri, skipVerify)

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
g := NewGraph(baseUri)

triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
g.Add(triple)

// w is of type io.Writer
g.Serialize(w, "text/turtle")
```

#### Using namespace prefixes

By default the Turtle serializer writes every IRI in full. Register namespace
prefixes on the graph with `AddPrefix` and the serializer will shorten matching
IRIs (including literal datatypes) and emit an `@prefix` header. Only the
prefixes actually used in the output are written. `AddPrefix` returns an error
if the label is not a valid Turtle prefix (for example if it contains a colon).

Output is deterministic: prefixes, subjects and predicate/object lines are
sorted, so the same graph always serializes to the same bytes.

```golang
g := NewGraph("https://example.org/foo")
g.AddPrefix("foaf", "http://xmlns.com/foaf/0.1/")
g.AddPrefix("xsd", "http://www.w3.org/2001/XMLSchema#")

me := NewResource("https://example.org/foo#me")
g.AddTriple(me, NewResource("http://xmlns.com/foaf/0.1/name"), NewLiteral("Alice"))
g.AddTriple(me, NewResource("http://xmlns.com/foaf/0.1/age"),
	NewLiteralWithDatatype("42", NewResource("http://www.w3.org/2001/XMLSchema#integer")))

g.Serialize(w, "text/turtle")
```

Output:

```turtle
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<https://example.org/foo#me>
  foaf:age "42"^^xsd:integer ;
  foaf:name "Alice" .
```

IRIs whose local part contains characters that cannot appear unescaped in a
prefixed name (for example `/` or a trailing `.`) are kept in full `<...>` form.
Prefixes are not inherited from parsed documents; add them explicitly.

### Serializing to JSON-LD

```golang
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g := NewGraph(baseUri)

triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
g.Add(triple)

// w is of type io.Writer
g.Serialize(w, "application/ld+json")
```
