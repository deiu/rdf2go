package rdf2go

import (
	"bytes"
	"strings"
	"testing"

	jsonld "github.com/linkeddata/gojsonld"
	"github.com/stretchr/testify/assert"
)

var (
	testUri    = "https://example.org"
	testTurtle = strings.NewReader("@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n<#me> a foaf:Person ;\nfoaf:name \"Test\".")
)

func TestNewGraph(t *testing.T) {
	g, err := NewGraph(testUri)
	assert.Nil(t, err)
	assert.Equal(t, testUri, g.URI())
	assert.Equal(t, 0, g.Len())
	assert.Equal(t, NewResource(testUri), g.Term())
}

func TestNewGraphError(t *testing.T) {
	_, err := NewGraph("ssh://test.org")
	assert.NotNil(t, err)
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

	t3 := NewLiteral("value")
	assert.True(t, t3.Equal(rdf2term(term2rdf(t3))))
	assert.True(t, t3.Equal(jterm2term(jsonld.NewLiteral("value"))))
}

func TestGraphBlankNodeTerms(t *testing.T) {
	t1 := NewBlankNode("n1")
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))
	assert.True(t, t1.Equal(jterm2term(jsonld.NewBlankNode("n1"))))
}

func TestParseFail(t *testing.T) {
	g, err := NewGraph("https://test.org/")
	assert.NoError(t, err)
	g.Parse(testTurtle, "text/n3")
	assert.Equal(t, 0, g.Len())
}

func TestParseTurtle(t *testing.T) {
	g, err := NewGraph("https://test.org/")
	assert.NoError(t, err)
	g.Parse(testTurtle, "text/turtle")
	assert.Equal(t, 2, g.Len())
}

func TestSerializeTurtle(t *testing.T) {
	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g, err := NewGraph("https://test.org/")
	assert.NoError(t, err)
	g.Add(triple)

	var b bytes.Buffer
	g.Serialize(&b, "text/turtle")
	assert.Equal(t, "<a>\n  <b> <c> .", b.String())

}

func TestParseJSONLD(t *testing.T) {
	data := "{ \"@id\": \"http://example.org/#me\", \"http://xmlns.com/foaf/0.1/name\": \"Test\" }"
	r := strings.NewReader(data)
	g, err := NewGraph("https://test.org/")
	assert.NoError(t, err)
	g.Parse(r, "application/ld+json")
	assert.Equal(t, 1, g.Len())
}

func TestSerializeJSONLD(t *testing.T) {
	triple := NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))
	g, err := NewGraph("https://test.org/")
	assert.NoError(t, err)
	g.Add(triple)

	var b bytes.Buffer
	g.Serialize(&b, "application/ld+json")
	assert.Equal(t, "{\"@id\":\"a\",\"b\":[{\"@id\":\"c\"}]}", b.String())
}
