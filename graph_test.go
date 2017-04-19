package rdf2go

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jsonld "github.com/linkeddata/gojsonld"
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
	g, err := NewGraph(testUri)
	assert.Nil(t, err)
	assert.Equal(t, testUri, g.URI())
	assert.Equal(t, 0, g.Len())
	assert.Equal(t, NewResource(testUri), g.Term())
}

// func TestNewGraphError(t *testing.T) {
// 	_, err := NewGraph("ssh://test.org")
// 	assert.NotNil(t, err)

// 	_, err = NewGraph("a")
// 	assert.NotNil(t, err)
// }

func TestGraphString(t *testing.T) {
	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g, err := NewGraph(testUri)
	assert.Nil(t, err)
	g.Add(triple)
	assert.Equal(t, "<a> <b> <c> .\n", g.String())
}

func TestGraphAdd(t *testing.T) {
	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g, err := NewGraph(testUri)
	assert.Nil(t, err)
	g.Add(triple)
	assert.Equal(t, 1, g.Len())
	g.Remove(triple)
	assert.Equal(t, 0, g.Len())
}

func TestGraphResourceTerms(t *testing.T) {
	t1 := NewResource(testUri)
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))

	t2 := NewResource(testUri)
	assert.True(t, t2.Equal(jterm2term(jsonld.NewResource(testUri))))
}

func TestGraphLiteralTerms(t *testing.T) {
	t1 := NewLiteralWithDatatype("value", NewResource(testUri))
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))
	assert.True(t, t1.Equal(jterm2term(jsonld.NewLiteralWithDatatype("value", jsonld.NewResource(testUri)))))

	t2 := NewLiteralWithLanguage("value", "en")
	assert.True(t, t2.Equal(rdf2term(term2rdf(t2))))
	assert.True(t, t2.Equal(jterm2term(jsonld.NewLiteralWithLanguage("value", "en"))))

	t3 := NewLiteral("value")
	assert.True(t, t3.Equal(rdf2term(term2rdf(t3))))
	assert.True(t, t3.Equal(jterm2term(jsonld.NewLiteral("value"))))
}

func TestGraphBlankNodeTerms(t *testing.T) {
	t1 := NewBlankNode("n1")
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))
	assert.True(t, t1.Equal(jterm2term(jsonld.NewBlankNode("n1"))))
}

func TestGraphOne(t *testing.T) {
	g, err := NewGraph(testUri)
	assert.NoError(t, err)

	assert.Nil(t, g.One(NewResource("a"), nil, nil))

	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g.Add(triple)

	assert.True(t, triple.Equal(g.One(NewResource("a"), NewResource("b"), NewResource("c"))))
	assert.True(t, triple.Equal(g.One(NewResource("a"), NewResource("b"), nil)))
	assert.True(t, triple.Equal(g.One(NewResource("a"), nil, nil)))

	assert.True(t, triple.Equal(g.One(nil, NewResource("b"), NewResource("c"))))
	assert.True(t, triple.Equal(g.One(nil, nil, NewResource("c"))))
	assert.True(t, triple.Equal(g.One(nil, NewResource("b"), nil)))

	assert.True(t, triple.Equal(g.One(nil, nil, nil)))
}

func TestGraphAll(t *testing.T) {
	g, err := NewGraph(testUri)
	assert.NoError(t, err)

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
	g, err := NewGraph(uri)
	assert.NoError(t, err)
	err = g.LoadURI(uri)
	assert.NoError(t, err)
	assert.Equal(t, 2, g.Len())
}

func TestGraphLoadURIFail(t *testing.T) {
	uri := testServer.URL + "/fail"
	g, err := NewGraph(uri)
	g.uri = ""
	assert.NoError(t, err)
	err = g.LoadURI(uri)
	assert.Error(t, err)
}

func TestParseFail(t *testing.T) {
	g, err := NewGraph(testUri)
	assert.NoError(t, err)
	g.Parse(strings.NewReader(simpleTurtle), "text/plain")
	assert.Equal(t, 0, g.Len())
}

func TestParseTurtle(t *testing.T) {
	g, err := NewGraph(testUri)
	assert.NoError(t, err)
	g.Parse(strings.NewReader(simpleTurtle), "text/turtle")
	assert.Equal(t, 2, g.Len())
}

func TestSerializeTurtle(t *testing.T) {
	triple1 := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g, err := NewGraph(testUri)
	assert.NoError(t, err)
	g.Add(triple1)

	b := new(bytes.Buffer)
	g.Serialize(b, "text/turtle")
	assert.Equal(t, "<a>\n  <b> <c> .", b.String())

	triple2 := NewTriple(NewResource("a"), NewResource("b"), NewResource("d"))
	g.Add(triple2)

	b = new(bytes.Buffer)
	g.Serialize(b, "text/turtle")
	toParse := strings.NewReader(b.String())
	g2, err := NewGraph(testUri)
	g2.Parse(toParse, "text/turtle")
	assert.Equal(t, 2, g2.Len())
}

func TestParseJSONLD(t *testing.T) {
	data := "{ \"@id\": \"http://example.org/#me\", \"http://xmlns.com/foaf/0.1/name\": \"Test\" }"
	r := strings.NewReader(data)
	g, err := NewGraph(testUri)
	assert.NoError(t, err)
	g.Parse(r, "application/ld+json")
	assert.Equal(t, 1, g.Len())
}

func TestSerializeJSONLD(t *testing.T) {
	g, err := NewGraph(testUri)
	assert.NoError(t, err)

	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g.Add(triple)
	triple = NewTriple(NewResource("a"), NewResource("d"), NewLiteral("e"))
	g.Add(triple)
	triple = NewTriple(NewResource("a"), NewResource("d"), NewLiteralWithDatatype("f", NewResource("g")))
	g.Add(triple)
	triple = NewTriple(NewResource("a"), NewResource("d"), NewLiteralWithLanguage("h", "en"))
	g.Add(triple)
	assert.Equal(t, 4, g.Len())

	// var b bytes.Buffer
	// g.Serialize(&b, "application/ld+json")
	// toParse := strings.NewReader(b.String())
	// g2, err := NewGraph(testUri)
	// g2.Parse(toParse, "text/turtle")
	// assert.Equal(t, 2, g2.Len())
}
