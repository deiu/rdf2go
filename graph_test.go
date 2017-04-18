package rdf2go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testUri = "https://example.org"
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

func TestGraphLiteralTerms(t *testing.T) {
	t1 := NewLiteralWithDatatype("value", NewResource(testUri))
	assert.True(t, t1.Equal(rdf2term(term2rdf(t1))))

	t2 := NewLiteralWithLanguage("value", "en")
	assert.True(t, t2.Equal(rdf2term(term2rdf(t2))))

	t3 := NewLiteral("value")
	assert.True(t, t3.Equal(rdf2term(term2rdf(t3))))
}
