package rdf2go

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	rdf "github.com/deiu/gon3"
	jsonld "github.com/linkeddata/gojsonld"
)

// Graph structure
type Graph struct {
	triples    map[*Triple]bool
	httpClient *http.Client
	uri        string
	term       Term
	prefixes   map[string]string
}

// NewHttpClient creates an http.Client to be used for parsing resources
// directly from the Web
func NewHttpClient(skip bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skip,
			},
		},
	}
}

// NewGraph creates a Graph object
func NewGraph(uri string, skipVerify ...bool) *Graph {
	skip := false
	if len(skipVerify) > 0 {
		skip = skipVerify[0]
	}
	g := &Graph{
		triples:    make(map[*Triple]bool),
		httpClient: NewHttpClient(skip),
		uri:        uri,
		term:       NewResource(uri),
		prefixes:   make(map[string]string),
	}
	return g
}

// AddPrefix registers a namespace prefix used by the Turtle serializer to
// shorten IRIs. The prefix is the short label without the trailing colon
// (e.g. "foaf") and the namespace is the full IRI it stands for
// (e.g. "http://xmlns.com/foaf/0.1/"). An empty prefix is allowed and
// produces names of the form ":local". It returns an error when the label
// cannot be written as a Turtle prefix or the namespace is empty.
func (g *Graph) AddPrefix(prefix string, namespace string) error {
	if !isValidPrefixLabel(prefix) {
		return fmt.Errorf("rdf2go: invalid prefix label %q", prefix)
	}
	if len(namespace) == 0 {
		return errors.New("rdf2go: prefix namespace must not be empty")
	}
	g.prefixes[prefix] = namespace
	return nil
}

// RemovePrefix drops a namespace prefix previously added with AddPrefix.
func (g *Graph) RemovePrefix(prefix string) {
	delete(g.prefixes, prefix)
}

// Prefixes returns a copy of the namespace prefixes registered on the graph.
func (g *Graph) Prefixes() map[string]string {
	out := make(map[string]string, len(g.prefixes))
	for k, v := range g.prefixes {
		out[k] = v
	}
	return out
}

// Len returns the length of the graph as number of triples in the graph
func (g *Graph) Len() int {
	return len(g.triples)
}

// Term returns a Graph Term object
func (g *Graph) Term() Term {
	return g.term
}

// URI returns a Graph URI object
func (g *Graph) URI() string {
	return g.uri
}

// One returns one triple based on a triple pattern of S, P, O objects
func (g *Graph) One(s Term, p Term, o Term) *Triple {
	for triple := range g.IterTriples() {
		if s != nil {
			if p != nil {
				if o != nil {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) && triple.Object.Equal(o) {
						return triple
					}
				} else {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) {
						return triple
					}
				}
			} else {
				if triple.Subject.Equal(s) {
					return triple
				}
			}
		} else if p != nil {
			if o != nil {
				if triple.Predicate.Equal(p) && triple.Object.Equal(o) {
					return triple
				}
			} else {
				if triple.Predicate.Equal(p) {
					return triple
				}
			}
		} else if o != nil {
			if triple.Object.Equal(o) {
				return triple
			}
		} else {
			return triple
		}
	}
	return nil
}

// IterTriples provides a channel containing all the triples in the graph.
// Note that the returned channel is already closed.
func (g *Graph) IterTriples() (ch chan *Triple) {
	// This function returns a channel rather than a slice for backwards compatibility.
	// It does not use a goroutine to populate the channel because that can trigger Go's 'concurrent map misuse'
	// detector, and would have little performance benefit.
	ch = make(chan *Triple, len(g.triples))
	for triple := range g.triples {
		ch <- triple
	}
	close(ch)
	return ch
}

// Add is used to add a Triple object to the graph
func (g *Graph) Add(t *Triple) {
	g.triples[t] = true
}

// AddTriple is used to add a triple made of individual S, P, O objects
func (g *Graph) AddTriple(s Term, p Term, o Term) {
	g.triples[NewTriple(s, p, o)] = true
}

// Remove is used to remove a Triple object
func (g *Graph) Remove(t *Triple) {
	delete(g.triples, t)
}

// All is used to return all triples that match a given pattern of S, P, O objects
func (g *Graph) All(s Term, p Term, o Term) []*Triple {
	var triples []*Triple
	for triple := range g.IterTriples() {
		if s != nil {
			if p != nil {
				if o != nil {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) && triple.Object.Equal(o) {
						triples = append(triples, triple)
					}
				} else {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) {
						triples = append(triples, triple)
					}
				}
			} else {
				if triple.Subject.Equal(s) {
					triples = append(triples, triple)
				}
			}
		} else if p != nil {
			if o != nil {
				if triple.Predicate.Equal(p) && triple.Object.Equal(o) {
					triples = append(triples, triple)
				}
			} else {
				if triple.Predicate.Equal(p) {
					triples = append(triples, triple)
				}
			}
		} else if o != nil {
			if triple.Object.Equal(o) {
				triples = append(triples, triple)
			}
		}
	}
	return triples
}

// Merge is used to add all the triples form another graph to this one
func (g *Graph) Merge(toMerge *Graph) {
	for triple := range toMerge.IterTriples() {
		g.Add(triple)
	}
}

// Parse is used to parse RDF data from a reader, using the provided mime type
func (g *Graph) Parse(reader io.Reader, mime string) error {
	parserName := mimeParser[normalizeMime(mime)]
	if len(parserName) == 0 {
		parserName = "guess"
	}
	if parserName == "jsonld" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		jsonData, err := jsonld.ReadJSON(buf.Bytes())
		if err != nil {
			return err
		}
		options := &jsonld.Options{}
		options.Base = ""
		options.ProduceGeneralizedRdf = false
		dataSet, err := jsonld.ToRDF(jsonData, options)
		if err != nil {
			return err
		}
		for t := range dataSet.IterTriples() {
			g.AddTriple(jterm2term(t.Subject), jterm2term(t.Predicate), jterm2term(t.Object))
		}

	} else if parserName == "turtle" {
		parser, err := rdf.NewParser(g.uri).Parse(reader)
		if err != nil {
			return err
		}
		for s := range parser.IterTriples() {
			g.AddTriple(rdf2term(s.Subject), rdf2term(s.Predicate), rdf2term(s.Object))
		}
	} else {
		return errors.New(parserName + " is not supported by the parser")
	}
	return nil
}

// LoadURI is used to load RDF data from a specific URI
func (g *Graph) LoadURI(uri string) error {
	doc := defrag(uri)
	q, err := http.NewRequest("GET", doc, nil)
	if err != nil {
		return err
	}
	if len(g.uri) == 0 {
		g.uri = doc
	}
	q.Header.Set("Accept", "text/turtle;q=1,application/ld+json;q=0.5")
	r, err := g.httpClient.Do(q)
	if err != nil {
		return err
	}
	if r != nil {
		defer r.Body.Close()
		if r.StatusCode == http.StatusOK {
			return g.Parse(r.Body, r.Header.Get("Content-Type"))
		}
		return fmt.Errorf("could not fetch graph from %s: HTTP %d", uri, r.StatusCode)
	}
	return nil
}

// String is used to serialize the graph object using NTriples
func (g *Graph) String() string {
	var toString string
	for triple := range g.IterTriples() {
		toString += triple.String() + "\n"
	}
	return toString
}

// Serialize is used to serialize a graph based on a given mime type
func (g *Graph) Serialize(w io.Writer, mime string) error {
	serializerName := mimeSerializer[normalizeMime(mime)]
	if serializerName == "jsonld" {
		return g.serializeJSONLD(w)
	}
	// just return Turtle by default
	return g.serializeTurtle(w)
}

// serializeTurtle writes the graph as Turtle. Output is deterministic:
// @prefix directives, subjects and predicate/object lines are all sorted.
func (g *Graph) serializeTurtle(w io.Writer) error {
	c := newCompacter(g.prefixes)

	// Encode everything first so that only the prefixes actually used end
	// up in the @prefix header.
	body := make(map[string][]string)
	for triple := range g.IterTriples() {
		s := c.encode(triple.Subject)
		body[s] = append(body[s], c.encode(triple.Predicate)+" "+c.encode(triple.Object))
	}

	subjects := make([]string, 0, len(body))
	for s := range body {
		subjects = append(subjects, s)
	}
	sort.Strings(subjects)

	for _, prefix := range c.usedPrefixes() {
		if _, err := fmt.Fprintf(w, "@prefix %s: <%s> .\n", prefix, g.prefixes[prefix]); err != nil {
			return err
		}
	}
	if len(c.used) > 0 {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	for _, subject := range subjects {
		if _, err := fmt.Fprintf(w, "%s\n", subject); err != nil {
			return err
		}
		lines := body[subject]
		sort.Strings(lines)
		for i, l := range lines {
			sep := " ;"
			if i == len(lines)-1 {
				sep = " ."
			}
			if _, err := fmt.Fprintf(w, "  %s%s\n", l, sep); err != nil {
				return err
			}
		}
	}

	return nil
}

// compacter shortens IRIs into prefixed names using a set of namespaces and
// remembers which prefixes it actually used. It is created per Serialize
// call so the bookkeeping never leaks between runs.
type compacter struct {
	// namespaces sorted by IRI length, longest first, so that the most
	// specific namespace wins when several share a common root.
	entries []prefixEntry
	used    map[string]bool
	cache   map[string]string
}

type prefixEntry struct {
	prefix    string
	namespace string
}

func newCompacter(prefixes map[string]string) *compacter {
	c := &compacter{
		used:  make(map[string]bool),
		cache: make(map[string]string),
	}
	for p, ns := range prefixes {
		if len(ns) == 0 || !isValidPrefixLabel(p) {
			continue
		}
		c.entries = append(c.entries, prefixEntry{prefix: p, namespace: ns})
	}
	sort.Slice(c.entries, func(i, j int) bool {
		if len(c.entries[i].namespace) != len(c.entries[j].namespace) {
			return len(c.entries[i].namespace) > len(c.entries[j].namespace)
		}
		return c.entries[i].prefix < c.entries[j].prefix
	})
	return c
}

// usedPrefixes returns the prefixes consumed so far, sorted alphabetically.
func (c *compacter) usedPrefixes() []string {
	out := make([]string, 0, len(c.used))
	for p := range c.used {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// compactIRI returns the prefixed form of an IRI when one of the registered
// namespaces matches and the remainder is a valid Turtle local name.
// Otherwise it returns the IRI wrapped in angle brackets.
func (c *compacter) compactIRI(iri string) string {
	if out, ok := c.cache[iri]; ok {
		return out
	}
	out := bracketIRI(iri)
	for _, e := range c.entries {
		if !strings.HasPrefix(iri, e.namespace) {
			continue
		}
		local := iri[len(e.namespace):]
		if !isValidLocalName(local) {
			continue
		}
		c.used[e.prefix] = true
		out = e.prefix + ":" + local
		break
	}
	c.cache[iri] = out
	return out
}

// encode renders a term in Turtle syntax, compacting IRIs where possible.
func (c *compacter) encode(iterm Term) string {
	return encodeTermWith(iterm, c.compactIRI)
}

// isValidPrefixLabel reports whether s can be used as the label of a Turtle
// @prefix directive. It accepts a conservative subset of the PN_PREFIX
// production: an empty string, or an ASCII letter followed by letters,
// digits, '_', '-' and '.', not ending with '.'.
func isValidPrefixLabel(s string) bool {
	if len(s) == 0 {
		return true
	}
	if !isASCIILetter(s[0]) {
		return false
	}
	return isValidNameTail(s[1:])
}

// isValidLocalName reports whether s can be written unescaped as the local
// part of a Turtle prefixed name. It accepts a conservative subset of the
// PN_LOCAL production: ASCII letters, digits, '_', '-' and '.', where the
// first character is a letter, digit or '_' and the name does not end with
// '.'. Any other character forces the full IRI form.
func isValidLocalName(s string) bool {
	if len(s) == 0 {
		return true
	}
	ch := s[0]
	if !isASCIILetter(ch) && !isASCIIDigit(ch) && ch != '_' {
		return false
	}
	return isValidNameTail(s[1:])
}

// isValidNameTail checks the characters allowed after the first one in a
// prefix label or local name.
func isValidNameTail(s string) bool {
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if !isASCIILetter(ch) && !isASCIIDigit(ch) && ch != '_' && ch != '-' && ch != '.' {
			return false
		}
	}
	return len(s) == 0 || s[len(s)-1] != '.'
}

func isASCIILetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// func (g *Graph) serializeJSONLD(w io.Writer) error {
// 	d := jsonld.NewDataset()
// 	triples := []*jsonld.Triple{}

// 	for triple := range g.IterTriples() {
// 		jTriple := jsonld.NewTriple(term2jterm(triple.Subject), term2jterm(triple.Predicate), term2jterm(triple.Object))
// 		triples = append(triples, jTriple)
// 	}

// 	d.Graphs[g.URI()] = triples
// 	opts := jsonld.NewOptions(g.URI())
// 	opts.UseNativeTypes = false
// 	opts.UseRdfType = true
// 	serializedJSON := jsonld.FromRDF(d, opts)
// 	jsonOut, err := json.MarshalIndent(serializedJSON, "", "    ")
// 	if err != nil {
// 		return err
// 	}

// 	_, err = fmt.Fprintf(w, "%s", jsonOut)
// 	return err
// }

func (g *Graph) serializeJSONLD(w io.Writer) error {
	r := []map[string]interface{}{}
	for elt := range g.IterTriples() {
		var one map[string]interface{}
		switch elt.Subject.(type) {
		case *BlankNode:
			one = map[string]interface{}{
				"@id": elt.Subject.(*BlankNode).String(),
			}
		default:
			one = map[string]interface{}{
				"@id": elt.Subject.(*Resource).URI,
			}
		}
		switch t := elt.Object.(type) {
		case *Resource:
			one[elt.Predicate.(*Resource).URI] = []map[string]string{
				{
					"@id": t.URI,
				},
			}
		case *Literal:
			v := map[string]string{
				"@value": t.Value,
			}
			if t.Datatype != nil && len(t.Datatype.String()) > 0 {
				v["@type"] = debrack(t.Datatype.String())
			}
			if len(t.Language) > 0 {
				v["@language"] = t.Language
			}
			one[elt.Predicate.(*Resource).URI] = []map[string]string{v}
		}
		r = append(r, one)
	}
	out, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
