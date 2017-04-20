package rdf2go

import (
	"github.com/stretchr/testify/assert"
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

func (tt *fakeTerm) RawValue() string {
	return ""
}

func TestTermResourceEqual(t *testing.T) {
	t1 := NewResource(testUri)
	assert.True(t, t1.Equal(NewResource(testUri)))
	assert.False(t, t1.Equal(NewLiteral("test1")))
}

func TestTermLiteral(t *testing.T) {
	str := "value"
	t1 := NewLiteral(str)
	assert.Equal(t, "\""+str+"\"", t1.String())
	assert.Equal(t, str, t1.RawValue())
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

func TestTermNewBlankNode(t *testing.T) {
	id := NewBlankNode(1)
	assert.Equal(t, "_:1", id.String())
	assert.Equal(t, "1", id.RawValue())
}

func TestTermNewAnonNode(t *testing.T) {
	id := NewAnonNode()
	assert.True(t, len(id.String()) > 1)
}

func TestTermBNodeEqual(t *testing.T) {
	id1 := NewBlankNode(1)
	id2 := NewBlankNode(1)
	assert.True(t, id1.Equal(id2))
	id3 := NewBlankNode(2)
	assert.False(t, id1.Equal(id3))
	assert.False(t, id1.Equal(NewResource(testUri)))
}

func TestTermNils(t *testing.T) {
	t1 := Term(&fakeTerm{URI: testUri})
	assert.Nil(t, term2rdf(t1))
	assert.Nil(t, term2jterm(t1))
}

func TestAtLang(t *testing.T) {
	assert.Equal(t, "@en", atLang("en"))
	assert.Equal(t, "@en", atLang("@en"))
	assert.Equal(t, "@e", atLang("e"))
	assert.Equal(t, "", atLang(""))
}

func TestEncodeTerm(t *testing.T) {
	iterm := NewResource(testUri)
	assert.Equal(t, iterm.String(), encodeTerm(iterm))
	iterm = NewBlankNode(1)
	assert.Equal(t, iterm.String(), encodeTerm(iterm))
	iterm = NewLiteral("value")
	assert.Equal(t, iterm.String(), encodeTerm(iterm))
	iterm = Term(&fakeTerm{URI: testUri})
	assert.Equal(t, "", encodeTerm(iterm))
}

func TestSplitPrefix(t *testing.T) {
	hashUri := testUri + "#me"
	base, name := splitPrefix(hashUri)
	assert.Equal(t, testUri+"#", base)
	assert.Equal(t, "me", name)

	slashUri := testUri + "/foaf"
	base, name = splitPrefix(slashUri)
	assert.Equal(t, testUri+"/", base)
	assert.Equal(t, "foaf", name)

	badUri := "test"
	base, name = splitPrefix(badUri)
	assert.Equal(t, "", base)
	assert.Equal(t, badUri, name)

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
