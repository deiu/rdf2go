package rdf2go

import (
	"regexp"
)

var mimeParser = map[string]string{
	"application/ld+json":       "jsonld",
	"application/json":          "internal",
	"application/sparql-update": "internal",
}

var mimeSerializer = map[string]string{
	"application/ld+json": "internal",
	"text/html":           "internal",
}

var mimeRdfExt = map[string]string{
	".ttl":    "text/turtle",
	".n3":     "text/n3",
	".rdf":    "application/rdf+xml",
	".jsonld": "application/ld+json",
}

var rdfExtensions = []string{
	".ttl",
	".n3",
	".rdf",
	".jsonld",
}

var (
	serializerMimes = []string{}
	validMimeType   = regexp.MustCompile(`^\w+/\w+$`)
)
