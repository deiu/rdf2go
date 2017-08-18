package rdf2go

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	rdf "github.com/deiu/gon3"
	jsonld "github.com/linkeddata/gojsonld"
)

// Graph structure
type Graph struct {
	triples    map[*Triple]bool
	httpClient *http.Client
	uri        string
	term       Term
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
	}
	return g
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

// IterTriples iterates through all the triples in a graph
func (g *Graph) IterTriples() (ch chan *Triple) {
	ch = make(chan *Triple)
	go func() {
		for triple := range g.triples {
			ch <- triple
		}
		close(ch)
	}()
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
func (g *Graph) Merge(toMerge *Graph){
	for triple := range toMerge.IterTriples(){
		g.Add(triple)
	}
}

// Parse is used to parse RDF data from a reader, using the provided mime type
func (g *Graph) Parse(reader io.Reader, mime string) error {
	parserName := mimeParser[mime]
	if len(parserName) == 0 {
		parserName = "guess"
	}
	if parserName == "jsonld" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		jsonData, err := jsonld.ReadJSON(buf.Bytes())
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
		if r.StatusCode == 200 {
			g.Parse(r.Body, r.Header.Get("Content-Type"))
		} else {
			return fmt.Errorf("Could not fetch graph from %s - HTTP %d", uri, r.StatusCode)
		}
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
	serializerName := mimeSerializer[mime]
	if serializerName == "jsonld" {
		return g.serializeJSONLD(w)
	}
	// just return Turtle by default
	return g.serializeTurtle(w)
}

// @TODO improve streaming
func (g *Graph) serializeTurtle(w io.Writer) error {
	var err error

	triplesBySubject := make(map[string][]*Triple)

	for triple := range g.IterTriples() {
		s := encodeTerm(triple.Subject)
		triplesBySubject[s] = append(triplesBySubject[s], triple)
	}

	for subject, triples := range triplesBySubject {
		_, err = fmt.Fprintf(w, "%s\n", subject)
		if err != nil {
			return err
		}

		for key, triple := range triples {
			p := encodeTerm(triple.Predicate)
			o := encodeTerm(triple.Object)

			if key == len(triples)-1 {
				_, err = fmt.Fprintf(w, "  %s %s .", p, o)
				if err != nil {
					return err
				}
				break
			}
			_, err = fmt.Fprintf(w, "  %s %s ;\n", p, o)
			if err != nil {
				return err
			}
		}

	}

	return nil
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
		one := map[string]interface{}{
			"@id": elt.Subject.(*Resource).URI,
		}
		switch t := elt.Object.(type) {
		case *Resource:
			one[elt.Predicate.(*Resource).URI] = []map[string]string{
				{
					"@id": t.URI,
				},
			}
			break
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
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, string(bytes))
	return nil
}
