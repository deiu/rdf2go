package rdf2go

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testServer *httptest.Server

	testUri      = "https://example.org"
	simpleTurtle = "@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n<#me> a foaf:Person ;\nfoaf:name \"Test\" ."
)

func init() {
	testServer = httptest.NewServer(MockServer())
	testServer.URL = strings.Replace(testServer.URL, "127.0.0.1", "localhost", 1)
}

func MockServer() http.Handler {
	// Create new handler
	handler := http.NewServeMux()
	handler.Handle("/foo", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/turtle")
		w.WriteHeader(200)
		w.Write([]byte(simpleTurtle))
	}))
	handler.Handle("/charset", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/turtle; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(simpleTurtle))
	}))
	return handler
}

func TestNewGraph(t *testing.T) {
	g := NewGraph(testUri)
	assert.Equal(t, testUri, g.URI())
	assert.Equal(t, 0, g.Len())
	assert.Equal(t, NewResource(testUri), g.Term())
}

func TestGraphString(t *testing.T) {
	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g := NewGraph(testUri)
	g.Add(triple)
	assert.Equal(t, "<a> <b> <c> .\n", g.String())
}

func TestGraphAdd(t *testing.T) {
	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g := NewGraph(testUri)
	g.Add(triple)
	assert.Equal(t, 1, g.Len())
	g.Remove(triple)
	assert.Equal(t, 0, g.Len())
}

func TestGraphResourceTerms(t *testing.T) {
	t1 := NewResource(testUri)
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))
	assert.True(t, t1.Equal(jterm2term(term2jterm(t1))))
}

func TestGraphLiteralTerms(t *testing.T) {
	t1 := NewLiteralWithDatatype("value", NewResource(testUri))
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))
	assert.True(t, t1.Equal(jterm2term(term2jterm(t1))))

	t2 := NewLiteralWithLanguage("value", "en")
	assert.True(t, t2.Equal(rdf2term(term2rdf(t2))))
	assert.True(t, t2.Equal(jterm2term(term2jterm(t2))))

	t3 := NewLiteral("value")
	assert.True(t, t3.Equal(rdf2term(term2rdf(t3))))
	assert.True(t, t3.Equal(jterm2term(term2jterm(t3))))
}

func TestGraphBlankNodeTerms(t *testing.T) {
	t1 := NewBlankNode("n1")
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))
	assert.True(t, t1.Equal(jterm2term(term2jterm(t1))))
}

func TestGraphOne(t *testing.T) {
	g := NewGraph(testUri)

	assert.Nil(t, g.One(NewResource("a"), nil, nil))

	triple := NewTriple(NewResource("a"), NewResource("foo#b"), NewResource("c"))
	g.Add(triple)

	assert.True(t, triple.Equal(g.One(NewResource("a"), NewResource("foo#b"), NewResource("c"))))
	assert.True(t, triple.Equal(g.One(NewResource("a"), NewResource("foo#b"), nil)))
	assert.True(t, triple.Equal(g.One(NewResource("a"), nil, nil)))

	assert.True(t, triple.Equal(g.One(nil, NewResource("foo#b"), NewResource("c"))))
	assert.True(t, triple.Equal(g.One(nil, nil, NewResource("c"))))
	assert.True(t, triple.Equal(g.One(nil, NewResource("foo#b"), nil)))

	assert.True(t, triple.Equal(g.One(nil, nil, nil)))
}

func TestGraphAll(t *testing.T) {
	g := NewGraph(testUri)

	assert.Empty(t, g.All(nil, nil, nil))

	g.AddTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g.AddTriple(NewResource("a"), NewResource("b"), NewResource("d"))
	g.AddTriple(NewResource("a"), NewResource("f"), NewLiteral("h"))
	g.AddTriple(NewResource("g"), NewResource("b2"), NewResource("e"))
	g.AddTriple(NewResource("g"), NewResource("b2"), NewResource("c"))

	assert.Equal(t, 0, len(g.All(nil, nil, nil)))
	assert.Equal(t, 3, len(g.All(NewResource("a"), nil, nil)))
	assert.Equal(t, 2, len(g.All(nil, NewResource("b"), nil)))
	assert.Equal(t, 1, len(g.All(nil, nil, NewResource("d"))))
	assert.Equal(t, 2, len(g.All(nil, nil, NewResource("c"))))
	assert.Equal(t, 1, len(g.All(NewResource("a"), NewResource("b"), NewResource("c"))))
	assert.Equal(t, 1, len(g.All(NewResource("a"), NewResource("f"), nil)))
	assert.Equal(t, 1, len(g.All(nil, NewResource("f"), NewLiteral("h"))))
}

func TestGraphLoadURI(t *testing.T) {
	uri := testServer.URL + "/foo#me"
	g := NewGraph(uri)
	err := g.LoadURI(uri)
	assert.NoError(t, err)
	assert.Equal(t, 2, g.Len())
}

func TestGraphLoadURIFail(t *testing.T) {
	uri := testServer.URL + "/fail"
	g := NewGraph(uri)
	g.uri = ""
	err := g.LoadURI(uri)
	assert.Error(t, err)
}

func TestGraphLoadURINoSkip(t *testing.T) {
	uri := testServer.URL + "/foo#me"
	g := NewGraph(uri, false)
	err := g.LoadURI(uri)
	assert.NoError(t, err)
	assert.Equal(t, 2, g.Len())
}

func TestParseFail(t *testing.T) {
	g := NewGraph(testUri)
	g.Parse(strings.NewReader(simpleTurtle), "text/plain")
	assert.Equal(t, 0, g.Len())
}

func TestParseTurtle(t *testing.T) {
	g := NewGraph(testUri)
	g.Parse(strings.NewReader(simpleTurtle), "text/turtle")
	assert.Equal(t, 2, g.Len())
	assert.NotNil(t, g.One(NewResource(testUri+"#me"), NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), NewResource("http://xmlns.com/foaf/0.1/Person")))
	assert.NotNil(t, g.One(NewResource(testUri+"#me"), NewResource("http://xmlns.com/foaf/0.1/name"), NewLiteral("Test")))

	prefixTurtle := "@prefix test: <http://example.org/test#> .\n<#me> test:foo \"Test\" ."
	g = NewGraph(testUri)
	g.Parse(strings.NewReader(prefixTurtle), "text/turtle")
	assert.Equal(t, 1, g.Len())
	assert.NotNil(t, g.One(NewResource(testUri+"#me"), NewResource("http://example.org/test#foo"), NewLiteral("Test")))
}

func TestSerializeTurtle(t *testing.T) {
	triple1 := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g := NewGraph(testUri)
	g.Add(triple1)

	b := new(bytes.Buffer)
	g.Serialize(b, "text/turtle")
	assert.Equal(t, "<a>\n  <b> <c> .\n", b.String())

	triple2 := NewTriple(NewResource("a"), NewResource("b"), NewResource("d"))
	g.Add(triple2)

	b = new(bytes.Buffer)
	g.Serialize(b, "text/turtle")
	toParse := strings.NewReader(b.String())
	g2 := NewGraph(testUri)
	g2.Parse(toParse, "text/turtle")
	assert.Equal(t, 2, g2.Len())
}

func TestSerializeTurtleMultipleSubjects(t *testing.T) {
	g := NewGraph(testUri)
	g.Add(NewTriple(NewResource("a"), NewResource("p"), NewResource("c")))
	g.Add(NewTriple(NewResource("b"), NewResource("p"), NewResource("d")))

	b := new(bytes.Buffer)
	g.Serialize(b, "text/turtle")
	out := b.String()
	// every subject block must end with " ." followed by a newline
	assert.Equal(t, 2, strings.Count(out, " .\n"))
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		assert.False(t, strings.Contains(line, ".<"), "statements run together on one line: %q", line)
	}
}

func TestSerializeTurtleWithPrefixes(t *testing.T) {
	g := NewGraph(testUri)
	g.AddPrefix("foaf", "http://xmlns.com/foaf/0.1/")
	g.AddPrefix("xsd", "http://www.w3.org/2001/XMLSchema#")
	g.AddPrefix("unused", "http://example.org/unused#")
	g.AddTriple(NewResource("http://example.org/#me"), NewResource("http://xmlns.com/foaf/0.1/name"), NewLiteral("Test"))
	g.AddTriple(NewResource("http://example.org/#me"), NewResource("http://xmlns.com/foaf/0.1/age"),
		NewLiteralWithDatatype("42", NewResource("http://www.w3.org/2001/XMLSchema#integer")))

	b := new(bytes.Buffer)
	assert.NoError(t, g.Serialize(b, "text/turtle"))
	out := b.String()

	// header lists only the prefixes that were used, sorted, followed by a blank line
	assert.True(t, strings.HasPrefix(out,
		"@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n\n"), out)
	assert.False(t, strings.Contains(out, "unused"), out)

	assert.Contains(t, out, "<http://example.org/#me>\n")
	assert.Contains(t, out, "foaf:name \"Test\"")
	assert.Contains(t, out, "foaf:age \"42\"^^xsd:integer")
	assert.False(t, strings.Contains(out, "<http://xmlns.com/foaf/0.1/name>"), out)

	// the output must parse back into an equal graph
	g2 := NewGraph(testUri)
	assert.NoError(t, g2.Parse(strings.NewReader(out), "text/turtle"))
	assert.Equal(t, g.Len(), g2.Len())
	for triple := range g.IterTriples() {
		assert.NotNil(t, g2.One(triple.Subject, triple.Predicate, triple.Object), triple.String())
	}
}

func TestLoadURIWithContentTypeParameters(t *testing.T) {
	g := NewGraph(testUri)
	assert.NoError(t, g.LoadURI(testServer.URL+"/charset"))
	assert.Equal(t, 2, g.Len())
}

func TestParseNormalizesMime(t *testing.T) {
	g := NewGraph(testUri)
	assert.NoError(t, g.Parse(strings.NewReader(simpleTurtle), "Text/Turtle; charset=UTF-8"))
	assert.Equal(t, 2, g.Len())
}

func TestSerializeTurtleIsDeterministic(t *testing.T) {
	g := NewGraph(testUri)
	g.AddTriple(NewResource("b"), NewResource("p"), NewResource("o"))
	g.AddTriple(NewResource("a"), NewResource("q"), NewResource("o"))
	g.AddTriple(NewResource("a"), NewResource("p"), NewResource("o2"))
	g.AddTriple(NewResource("a"), NewResource("p"), NewResource("o1"))

	want := "<a>\n  <p> <o1> ;\n  <p> <o2> ;\n  <q> <o> .\n<b>\n  <p> <o> .\n"
	for i := 0; i < 20; i++ {
		b := new(bytes.Buffer)
		assert.NoError(t, g.Serialize(b, "text/turtle"))
		assert.Equal(t, want, b.String())
	}
}

func TestAddPrefixRejectsInvalidLabels(t *testing.T) {
	g := NewGraph(testUri)
	assert.NoError(t, g.AddPrefix("", "http://example.org/"))
	assert.NoError(t, g.AddPrefix("ex", "http://example.org/"))
	assert.NoError(t, g.AddPrefix("a-b.c1", "http://example.org/"))
	assert.Error(t, g.AddPrefix("ex:", "http://example.org/"))
	assert.Error(t, g.AddPrefix("_", "http://example.org/"))
	assert.Error(t, g.AddPrefix("1x", "http://example.org/"))
	assert.Error(t, g.AddPrefix("a b", "http://example.org/"))
	assert.Error(t, g.AddPrefix("a.", "http://example.org/"))
	assert.Error(t, g.AddPrefix("ok", ""))
	assert.Len(t, g.Prefixes(), 3)
}

func TestSerializeTurtleEmptyPrefix(t *testing.T) {
	g := NewGraph(testUri)
	assert.NoError(t, g.AddPrefix("", "http://example.org/"))
	g.AddTriple(NewResource("http://example.org/s"), NewResource("http://example.org/p"), NewResource("http://example.org/o"))
	b := new(bytes.Buffer)
	assert.NoError(t, g.Serialize(b, "text/turtle"))
	assert.Equal(t, "@prefix : <http://example.org/> .\n\n:s\n  :p :o .\n", b.String())

	g2 := NewGraph(testUri)
	assert.NoError(t, g2.Parse(strings.NewReader(b.String()), "text/turtle"))
	assert.Equal(t, 1, g2.Len())
}

// customTerm is a user-defined Term implementation used to check that the
// serializer does not assume datatypes are *Resource values.
type customTerm struct{ iri string }

func (c *customTerm) String() string    { return "<" + c.iri + ">" }
func (c *customTerm) RawValue() string  { return c.iri }
func (c *customTerm) Equal(o Term) bool { return false }

func TestSerializeTurtleCustomDatatypeTerm(t *testing.T) {
	g := NewGraph(testUri)
	g.AddPrefix("xsd", "http://www.w3.org/2001/XMLSchema#")
	g.AddTriple(NewResource("s"), NewResource("p"), NewLiteralWithDatatype("42", &customTerm{iri: "http://www.w3.org/2001/XMLSchema#integer"}))
	b := new(bytes.Buffer)
	assert.NoError(t, g.Serialize(b, "text/turtle"))
	// a datatype that is not a *Resource falls back to its own String()
	assert.Contains(t, b.String(), "\"42\"^^<http://www.w3.org/2001/XMLSchema#integer>")
	assert.False(t, strings.Contains(b.String(), "@prefix"), b.String())
}

func TestSerializeTurtleNoPrefixesNoHeader(t *testing.T) {
	g := NewGraph(testUri)
	g.AddTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	b := new(bytes.Buffer)
	assert.NoError(t, g.Serialize(b, "text/turtle"))
	assert.Equal(t, "<a>\n  <b> <c> .\n", b.String())
}

func TestSerializeTurtlePrefixFallsBackToFullIRI(t *testing.T) {
	g := NewGraph(testUri)
	g.AddPrefix("ex", "http://example.org/")
	// local names with characters outside the safe subset stay as full IRIs
	g.AddTriple(NewResource("http://example.org/a/b"), NewResource("http://example.org/p"), NewResource("http://example.org/trailing."))
	g.AddTriple(NewResource("http://example.org/-dash"), NewResource("http://example.org/p"), NewResource("http://example.org/"))

	b := new(bytes.Buffer)
	assert.NoError(t, g.Serialize(b, "text/turtle"))
	out := b.String()
	assert.Contains(t, out, "<http://example.org/a/b>")
	assert.Contains(t, out, "<http://example.org/trailing.>")
	assert.Contains(t, out, "<http://example.org/-dash>")
	assert.Contains(t, out, "ex:p")
	// empty local name is allowed
	assert.Contains(t, out, " ex: .")

	g2 := NewGraph(testUri)
	assert.NoError(t, g2.Parse(strings.NewReader(out), "text/turtle"))
	assert.Equal(t, 2, g2.Len())
}

func TestSerializeTurtleLongestPrefixWins(t *testing.T) {
	g := NewGraph(testUri)
	g.AddPrefix("ex", "http://example.org/")
	g.AddPrefix("exv", "http://example.org/vocab#")
	g.AddTriple(NewResource("http://example.org/vocab#Thing"), NewResource("http://example.org/p"), NewResource("http://example.org/o"))

	b := new(bytes.Buffer)
	assert.NoError(t, g.Serialize(b, "text/turtle"))
	assert.Contains(t, b.String(), "exv:Thing")
	assert.Contains(t, b.String(), "ex:p ex:o")
}

func TestPrefixes(t *testing.T) {
	g := NewGraph(testUri)
	g.AddPrefix("ex", "http://example.org/")
	assert.Equal(t, map[string]string{"ex": "http://example.org/"}, g.Prefixes())
	g.Prefixes()["ex"] = "mutated"
	assert.Equal(t, "http://example.org/", g.Prefixes()["ex"])
	g.RemovePrefix("ex")
	assert.Empty(t, g.Prefixes())
}

func TestIsValidPrefixLabel(t *testing.T) {
	assert.True(t, isValidPrefixLabel(""))
	assert.True(t, isValidPrefixLabel("ex"))
	assert.True(t, isValidPrefixLabel("a-b.c1"))
	assert.False(t, isValidPrefixLabel("1x"))
	assert.False(t, isValidPrefixLabel("_x"))
	assert.False(t, isValidPrefixLabel("ex:"))
	assert.False(t, isValidPrefixLabel("a."))
	assert.False(t, isValidPrefixLabel("a b"))
}

func TestIsValidLocalName(t *testing.T) {
	assert.True(t, isValidLocalName(""))
	assert.True(t, isValidLocalName("name"))
	assert.True(t, isValidLocalName("a-b.c_1"))
	assert.False(t, isValidLocalName("-a"))
	assert.False(t, isValidLocalName(".a"))
	assert.False(t, isValidLocalName("a."))
	assert.False(t, isValidLocalName("a/b"))
	assert.False(t, isValidLocalName("a:b"))
	assert.False(t, isValidLocalName("é"))
}

func TestParseJSONLD(t *testing.T) {
	data := "{ \"@id\": \"http://example.org/#me\", \"http://xmlns.com/foaf/0.1/name\": \"Test\" }"
	r := strings.NewReader(data)
	g := NewGraph(testUri)
	g.Parse(r, "application/ld+json")
	assert.Equal(t, 1, g.Len())
}

func TestSerializeJSONLD(t *testing.T) {
	g := NewGraph(testUri)
	g.Parse(strings.NewReader(simpleTurtle), "text/turtle")
	g.Add(NewTriple(NewResource(testUri+"#me"), NewResource("http://xmlns.com/foaf/0.1/nick"), NewLiteralWithLanguage("test", "en")))
	g.Add(NewTriple(NewBlankNode("n9"), NewResource("http://xmlns.com/foaf/0.1/name"), NewLiteralWithLanguage("test", "en")))
	assert.Equal(t, 4, g.Len())

	var b bytes.Buffer
	g.Serialize(&b, "application/ld+json")
	toParse := strings.NewReader(b.String())
	g2 := NewGraph(testUri)
	g2.Parse(toParse, "application/ld+json")
	assert.Equal(t, 4, g2.Len())
}

func TestGraphMerge(t *testing.T) {
	g := NewGraph(testUri)
	g2 := NewGraph(testUri)

	g.AddTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g.AddTriple(NewResource("a"), NewResource("b"), NewResource("d"))
	g.AddTriple(NewResource("a"), NewResource("f"), NewLiteral("h"))
	assert.Equal(t, 3, g.Len())
	g2.AddTriple(NewResource("g"), NewResource("b2"), NewResource("e"))
	g2.AddTriple(NewResource("g"), NewResource("b2"), NewResource("c"))
	assert.Equal(t, 2, g2.Len())

	g.Merge(g2)

	assert.Equal(t, 5, g.Len())
	assert.NotEqual(t, nil, g.One(NewResource("a"), NewResource("b"), NewResource("c")))
	assert.NotEqual(t, nil, g.One(NewResource("a"), NewResource("b"), NewResource("d")))
	assert.NotEqual(t, nil, g.One(NewResource("a"), NewResource("f"), NewResource("h")))
	assert.NotEqual(t, nil, g.One(NewResource("g"), NewResource("b2"), NewResource("e")))
	assert.NotEqual(t, nil, g.One(NewResource("g"), NewResource("b2"), NewResource("c")))
}
