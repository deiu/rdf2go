package rdf2go

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	rdf "github.com/deiu/gon3"
	jsonld "github.com/linkeddata/gojsonld"
)

// AnyGraph defines methods common to Graph types
// type AnyGraph interface {
// 	Len() int
// 	URI() string
// 	Parse(io.Reader, string)
// 	Serialize(string) (string, error)

// 	IterTriples() chan *Triple

// 	ReadFile(string)
// 	WriteFile(*os.File, string) error
// }

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

// Graph structure
type Graph struct {
	triples map[*Triple]bool

	uri  string
	term Term
}

// NewGraph creates a Graph object
func NewGraph(uri string) (*Graph, error) {
	if uri[:5] != "http:" && uri[:6] != "https:" {
		return &Graph{}, errors.New("Non http graphs are not allowed")
	}
	return &Graph{
		triples: make(map[*Triple]bool),

		uri:  uri,
		term: NewResource(uri),
	}, nil
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

func term2rdf(t Term) rdf.Term {
	switch t := t.(type) {
	case *BlankNode:
		id := t.RawValue()
		node := rdf.NewBlankNode(id)
		return node
	case *Resource:
		node, _ := rdf.NewIRI(t.RawValue())
		return node
	case *Literal:
		if t.Datatype != nil {
			iri, _ := rdf.NewIRI(t.Datatype.(*Resource).URI)
			return rdf.NewLiteralWithDataType(t.Value, iri)
		}
		if len(t.Language) > 0 {
			node := rdf.NewLiteralWithLanguage(t.Value, t.Language)
			return node
		}
		node := rdf.NewLiteral(t.Value)
		return node
	}
	return nil
}

func rdf2term(term rdf.Term) Term {
	switch term := term.(type) {
	case *rdf.BlankNode:
		return NewBlankNode(term.RawValue())
	case *rdf.Literal:
		if len(term.LanguageTag) > 0 {
			return NewLiteralWithLanguage(term.LexicalForm, term.LanguageTag)
		}
		if term.DatatypeIRI != nil && len(term.DatatypeIRI.String()) > 0 {
			return NewLiteralWithLanguageAndDatatype(term.LexicalForm, term.LanguageTag, NewResource(debrack(term.DatatypeIRI.String())))
		}
		return NewLiteral(term.RawValue())
	case *rdf.IRI:
		return NewResource(term.RawValue())
	}
	return nil
}

func jterm2term(term jsonld.Term) Term {
	switch term := term.(type) {
	case *jsonld.BlankNode:
		return NewBlankNode(term.RawValue())
	case *jsonld.Literal:
		if term.Datatype != nil && len(term.Datatype.String()) > 0 {
			return NewLiteralWithLanguageAndDatatype(term.Value, term.Language, NewResource(term.Datatype.RawValue()))
		}
		return NewLiteral(term.Value)
	case *jsonld.Resource:
		return NewResource(term.RawValue())
	}
	return nil
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

// AddStatement adds a Statement object
// func (g *Graph) AddStatement(st *crdf.Statement) {
// 	s, p, o := term2term(st.Subject), term2term(st.Predicate), term2term(st.Object)
// 	for range g.All(s, p, o) {
// 		return
// 	}
// 	g.AddTriple(s, p, o)
// }

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

// ParseBase is used to parse RDF data from a reader, using the provided mime type and a base URI
// func (g *Graph) ParseBase(reader io.Reader, mime string, baseURI string) {
// 	if len(baseURI) < 1 {
// 		baseURI = g.uri
// 	}
// 	parserName := mimeParser[mime]
// 	if len(parserName) == 0 {
// 		parserName = "guess"
// 	}
// 	parser := crdf.NewParser(parserName)
// 	defer parser.Free()
// 	out := parser.Parse(reader, baseURI)
// 	for s := range out {
// 		g.AddStatement(s)
// 	}
// }

// ReadFile is used to read RDF data from a file into the graph
func (g *Graph) ReadFile(filename string) {
	stat, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return
	} else if stat.IsDir() {
		return
	} else if !stat.IsDir() && err != nil {
		log.Println(err)
		return
	}
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	defer f.Close()
	if err != nil {
		log.Println(err)
		return
	}
	g.Parse(f, "text/turtle")
}

// LoadURI is used to load RDF data from a specific URI
func (g *Graph) LoadURI(uri string) (err error) {
	doc := defrag(uri)
	q, err := http.NewRequest("GET", doc, nil)
	if err != nil {
		return
	}
	if len(g.uri) == 0 {
		g.uri = doc
	}
	q.Header.Set("Accept", "text/turtle,text/n3,application/rdf+xml")
	r, err := httpClient.Do(q)
	if err != nil {
		return
	}
	if r != nil {
		defer r.Body.Close()
		if r.StatusCode == 200 {
			g.Parse(r.Body, r.Header.Get("Content-Type"))
		} else {
			err = fmt.Errorf("Could not fetch graph from %s - HTTP %d", uri, r.StatusCode)
		}
	}
	return
}

func (g *Graph) serializeTurtle(w io.Writer) error {
	var err error

	triplesBySubject := make(map[string][]*Triple)

	for triple := range g.IterTriples() {
		s := encodeTerm(triple.Subject)
		triplesBySubject[s] = append(triplesBySubject[s], triple)
	}

	_, err = fmt.Fprint(w, "\n")
	if err != nil {
		return err
	}

	for subject, triples := range triplesBySubject {
		_, err = fmt.Fprintf(w, "%s\n", subject)
		if err != nil {
			return err
		}

		for _, triple := range triples {
			p := encodeTerm(triple.Predicate)
			o := encodeTerm(triple.Object)

			_, err = fmt.Fprintf(w, "  %s %s ;\n", p, o)
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprintf(w, "  .\n\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) serializeJSONLd(w io.Writer) error {
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
				v["@type"] = t.Datatype.String()
			}
			if len(t.Language) > 0 {
				v["@language"] = t.Language
			}
			one[elt.Predicate.(*Resource).URI] = []map[string]string{v}
		}
		r = append(r, one)
	}
	tree, err := json.Marshal(r)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, string(tree))
	return nil
}

// Serialize is used to serialize a graph based on a given mime type
func (g *Graph) Serialize(w io.Writer, mime string) error {
	if mime == "application/ld+json" {
		err := g.serializeJSONLd(w)
		return err
	}

	serializerName := mimeSerializer[mime]
	if len(serializerName) == 0 {
		serializerName = "turtle"
	}
	err := g.serializeTurtle(w)
	return err
}

// WriteFile is used to dump RDF from a Graph into a file
// func (g *Graph) WriteFile(file *os.File, mime string) error {
// 	serializerName := mimeSerializer[mime]
// 	if len(serializerName) == 0 {
// 		serializerName = "turtle"
// 	}
// 	serializer := crdf.NewSerializer(serializerName)
// 	defer serializer.Free()
// 	err := serializer.SetFile(file, g.uri)
// 	if err != nil {
// 		return err
// 	}
// 	ch := make(chan *crdf.Statement, 1024)
// 	go func() {
// 		for triple := range g.IterTriples() {
// 			ch <- &crdf.Statement{
// 				Subject:   term2C(triple.Subject),
// 				Predicate: term2C(triple.Predicate),
// 				Object:    term2C(triple.Object),
// 			}
// 		}
// 		close(ch)
// 	}()
// 	serializer.AddN(ch)
// 	return nil
// }

func encodeTerm(iterm Term) string {
	switch term := iterm.(type) {
	case *Resource:
		return fmt.Sprintf("<%s>", term.URI)
	case *Literal:
		return term.String()
	case *BlankNode:
		return term.String()
	}

	return ""
}

// splitPrefix takes a given URI and splits it into a base URI and a local name
func splitPrefix(uri string) (base string, name string) {
	index := strings.LastIndex(uri, "#") + 1

	if index > 0 {
		return uri[:index], uri[index:]
	}

	index = strings.LastIndex(uri, "/") + 1

	if index > 0 {
		return uri[:index], uri[index:]
	}

	return "", uri
}

func brack(s string) string {
	if len(s) > 0 && s[0] == '<' {
		return s
	}
	if len(s) > 0 && s[len(s)-1] == '>' {
		return s
	}
	return "<" + s + ">"
}

func debrack(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] != '<' {
		return s
	}
	if s[len(s)-1] != '>' {
		return s
	}
	return s[1 : len(s)-1]
}

func defrag(s string) string {
	lst := strings.Split(s, "#")
	if len(lst) != 2 {
		return s
	}
	return lst[0]
}
