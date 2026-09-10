// Package rdf2go is a small RDF library for Go.
//
// It provides an in-memory Graph of Triples built from Resource (IRI),
// Literal and BlankNode terms, together with parsers and serializers for
// Turtle (text/turtle) and JSON-LD (application/ld+json). Graphs can also be
// loaded directly from the Web with Graph.LoadURI.
//
// The Turtle serializer writes full IRIs by default. Register namespaces with
// Graph.AddPrefix to have matching IRIs shortened to prefixed names and an
// @prefix header emitted.
package rdf2go
