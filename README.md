# rdf2go
Native golang parser/serializer from/to Turtle and JSONLD.

# Installation

Just go get it!

`go get -u github.com/deiu/rdf2go`

# Example usage

## Working with graphs

```
// Set a base URI
baseUri := "https://example.org/foo"

// Create a new graph
g, err := NewGraph(baseUri)
if err != nil {
	// deal with err
}

// Add a new triple to the graph
triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
g.Add(triple)

// Get length of Graph (nr of triples)
g.Len()

// Delete the triple
g.Remove(triple)
```

## Looking up triples from the graph

### Looking up a single triple

The `g.One()` method returns the first triple that matches against any (or all) of Subject, Predicate, Object patterns.

```
// Create a new graph
g, _ := NewGraph("https://example.org")

// Add a few triples
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("c")))

// Look up one triple matching the given subject
triple := g.One(nil, nil, NewResource("c"))
triple.String() // -> <a> <b> <c> .

triple = g.One(nil, NewResource("b"), nil)
triple.String() // -> <a> <b> <c> .

triple = g.One(nil, NewResource("z"), nil) // -> nil
```

### Looking up triples and returning a list of matches

Similar to `g.One()`, `g.All()` returns all triples that match the given pattern.

```
// Create a new graph
g, _ := NewGraph("https://example.org")

// Add a few triples
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("c")))
g.Add(NewTriple(NewResource("a"), NewResource("b"), NewResource("d")))

// Look up one triple matching the given subject
triples := g.All(nil, nil, NewResource("c")) //
for triple := range triples {
	triple.String() // -> <a> <b> <c> .
}	

triples = g.All(nil, NewResource("b"), nil)
for triple := range triples {
	triple.String()
}
// Prints: 
// <a> <b> <c> .
// <a> <b> <d> .
```

## Different types of terms (resources)

### IRIs

```
// Create a new IRI
iri := NewResource("https://example.org")
iri.String() // -> <https://example.org>
```

### Literals

```
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

```
// Create a new Blank Node
bn := NewBlankNode("a1")
bn.String() // -> "_:a1"
```