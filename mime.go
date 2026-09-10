package rdf2go

import "mime"

// normalizeMime strips parameters such as "; charset=utf-8" from a media
// type and lowercases it so it can be used as a lookup key. Unparseable
// values are returned unchanged.
func normalizeMime(value string) string {
	if mt, _, err := mime.ParseMediaType(value); err == nil {
		return mt
	}
	return value
}

var mimeParser = map[string]string{
	"text/turtle":               "turtle",
	"application/ld+json":       "jsonld",
	"application/sparql-update": "internal",
}

var mimeSerializer = map[string]string{
	"application/ld+json": "jsonld",
	"text/html":           "internal",
}
