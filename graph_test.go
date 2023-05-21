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
		return
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
	assert.Equal(t, "<a>\n  <b> <c> .", b.String())

	triple2 := NewTriple(NewResource("a"), NewResource("b"), NewResource("d"))
	g.Add(triple2)

	b = new(bytes.Buffer)
	g.Serialize(b, "text/turtle")
	toParse := strings.NewReader(b.String())
	g2 := NewGraph(testUri)
	g2.Parse(toParse, "text/turtle")
	assert.Equal(t, 2, g2.Len())
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
	assert.Equal(t,3,g.Len())
	g2.AddTriple(NewResource("g"), NewResource("b2"), NewResource("e"))
	g2.AddTriple(NewResource("g"), NewResource("b2"), NewResource("c"))
	assert.Equal(t,2,g2.Len())

	g.Merge(g2)

	assert.Equal(t,5,g.Len())
    assert.NotEqual(t,nil,g.One(NewResource("a"),NewResource("b"),NewResource("c")))
	assert.NotEqual(t,nil,g.One(NewResource("a"),NewResource("b"),NewResource("d")))
	assert.NotEqual(t,nil,g.One(NewResource("a"),NewResource("f"),NewResource("h")))
	assert.NotEqual(t,nil,g.One(NewResource("g"),NewResource("b2"),NewResource("e")))
	assert.NotEqual(t,nil,g.One(NewResource("g"),NewResource("b2"),NewResource("c")))
}
