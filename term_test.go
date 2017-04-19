package rdf2go

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type fakeTerm struct {
	URI string
}

func (tt *fakeTerm) Equal(term Term) bool {
	return false
}

func (tt *fakeTerm) String() string {
	return ""
}

func TestTermResourceEqual(t *testing.T) {
	t1 := NewResource(testUri)
	assert.True(t, t1.Equal(NewResource(testUri)))
	assert.False(t, t1.Equal(NewLiteral("test1")))
}

func TestTermLiteralEqual(t *testing.T) {
	t1 := NewLiteralWithLanguage("test1", "en")
	assert.False(t, t1.Equal(NewResource(testUri)))

	assert.True(t, t1.Equal(NewLiteralWithLanguage("test1", "en")))
	assert.False(t, t1.Equal(NewLiteralWithLanguage("test2", "en")))

	assert.True(t, t1.Equal(NewLiteralWithLanguage("test1", "en")))
	assert.False(t, t1.Equal(NewLiteralWithLanguage("test1", "fr")))

	t1 = NewLiteralWithDatatype("test1", NewResource("http://www.w3.org/2001/XMLSchema#string"))
	assert.False(t, t1.Equal(NewLiteral("test1")))
	assert.True(t, t1.Equal(NewLiteralWithDatatype("test1", NewResource("http://www.w3.org/2001/XMLSchema#string"))))
	assert.False(t, t1.Equal(NewLiteralWithDatatype("test1", NewResource("http://www.w3.org/2001/XMLSchema#int"))))
}

func TestTermNewLiteralWithLanguage(t *testing.T) {
	s := NewLiteralWithLanguage("test", "en")
	assert.Equal(t, "\"test\"@en", s.String())
}

func TestTermNewLiteralWithDatatype(t *testing.T) {
	s := NewLiteralWithDatatype("test", NewResource("http://www.w3.org/2001/XMLSchema#string"))
	assert.Equal(t, "\"test\"^^<http://www.w3.org/2001/XMLSchema#string>", s.String())
}

func TestTermNewLiteralWithLanguageAndDatatype(t *testing.T) {
	s := NewLiteralWithLanguageAndDatatype("test", "en", NewResource("http://www.w3.org/2001/XMLSchema#string"))
	assert.Equal(t, "\"test\"@en", s.String())

	s = NewLiteralWithLanguageAndDatatype("test", "", NewResource("http://www.w3.org/2001/XMLSchema#string"))
	assert.Equal(t, "\"test\"^^<http://www.w3.org/2001/XMLSchema#string>", s.String())
}

func TestTermNewBlankNode(t *testing.T) {
	id := NewBlankNode("n1")
	assert.Equal(t, "_:n1", id.String())
}

func TestTermNewAnonNode(t *testing.T) {
	id := NewAnonNode()
	assert.True(t, strings.Contains(id.String(), "_:anon"))
}

func TestTermBNodeEqual(t *testing.T) {
	id1 := NewBlankNode("n1")
	id2 := NewBlankNode("n1")
	assert.True(t, id1.Equal(id2))
	id3 := NewBlankNode("n2")
	assert.False(t, id1.Equal(id3))
	assert.False(t, id1.Equal(NewResource(testUri)))
}

func TestTermNils(t *testing.T) {
	t1 := Term(&fakeTerm{URI: "test"})
	assert.Nil(t, term2rdf(t1))
}

func TestRDFBrack(t *testing.T) {
	assert.Equal(t, "<test>", brack("test"))
	assert.Equal(t, "<test", brack("<test"))
	assert.Equal(t, "test>", brack("test>"))
}

func TestRDFDebrack(t *testing.T) {
	assert.Equal(t, "a", debrack("a"))
	assert.Equal(t, "test", debrack("<test>"))
	assert.Equal(t, "<test", debrack("<test"))
	assert.Equal(t, "test>", debrack("test>"))
}

func TestDefrag(t *testing.T) {
	assert.Equal(t, "test", defrag("test"))
	assert.Equal(t, "test", defrag("test#me"))
}
