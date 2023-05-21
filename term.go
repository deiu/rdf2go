/*
	Copyright (c) 2012 Kier Davis

	Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
	associated documentation files (the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute,
	sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all copies or substantial
	portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
	NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
	NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES
	OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
	CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package rdf2go

import (
	"fmt"
	"math/rand"
	"strings"

	rdf "github.com/deiu/gon3"
	jsonld "github.com/linkeddata/gojsonld"
)

// A Term is the value of a subject, predicate or object i.e. a IRI reference, blank node or
// literal.
type Term interface {
	// Method String should return the NTriples representation of this term.
	String() string

	// Method RawValue should return the raw value of this term.
	RawValue() string

	// Method Equal should return whether this term is equal to another.
	Equal(Term) bool
}

// Resource is an URI / IRI reference.
type Resource struct {
	URI string
}

// NewResource returns a new resource object.
func NewResource(uri string) (term Term) {
	return Term(&Resource{URI: uri})
}

// String returns the NTriples representation of this resource.
func (term Resource) String() (str string) {
	return fmt.Sprintf("<%s>", term.URI)
}

// RawValue returns the string value of the a resource without brackets.
func (term Resource) RawValue() (str string) {
	return term.URI
}

// Equal returns whether this resource is equal to another.
func (term Resource) Equal(other Term) bool {
	if spec, ok := other.(*Resource); ok {
		return term.URI == spec.URI
	}

	return false
}

// Literal is a textual value, with an associated language or datatype.
type Literal struct {
	Value    string
	Language string
	Datatype Term
}

// NewLiteral returns a new literal with the given value.
func NewLiteral(value string) (term Term) {
	return Term(&Literal{Value: value})
}

// NewLiteralWithLanguage returns a new literal with the given value and language.
func NewLiteralWithLanguage(value string, language string) (term Term) {
	return Term(&Literal{Value: value, Language: language})
}

// NewLiteralWithDatatype returns a new literal with the given value and datatype.
func NewLiteralWithDatatype(value string, datatype Term) (term Term) {
	return Term(&Literal{Value: value, Datatype: datatype})
}

// String returns the NTriples representation of this literal.
func (term Literal) String() string {
	str := term.Value
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	str = strings.Replace(str, "\n", "\\n", -1)
	str = strings.Replace(str, "\r", "\\r", -1)
	str = strings.Replace(str, "\t", "\\t", -1)

	str = fmt.Sprintf("\"%s\"", str)

	// if term.Language != "" {
	str += atLang(term.Language)
	// } else
	if term.Datatype != nil {
		str += "^^" + term.Datatype.String()
	}

	return str
}

func (term Literal) RawValue() string {
	return term.Value
}

// Equal returns whether this literal is equivalent to another.
func (term Literal) Equal(other Term) bool {
	spec, ok := other.(*Literal)
	if !ok {
		return false
	}

	if term.Value != spec.Value {
		return false
	}

	if term.Language != spec.Language {
		return false
	}

	if (term.Datatype == nil && spec.Datatype != nil) || (term.Datatype != nil && spec.Datatype == nil) {
		return false
	}

	if term.Datatype != nil && spec.Datatype != nil && !term.Datatype.Equal(spec.Datatype) {
		return false
	}

	return true
}

// BlankNode is an RDF blank node i.e. an unqualified URI/IRI.
type BlankNode struct {
	ID string
}

// NewBlankNode returns a new blank node with the given ID.
func NewBlankNode(id string) (term Term) {
	return Term(&BlankNode{ID: id})
}

// NewAnonNode returns a new blank node with a pseudo-randomly generated ID.
func NewAnonNode() (term Term) {
	return Term(&BlankNode{ID: fmt.Sprint("n", rand.Int())})
}

// String returns the NTriples representation of the blank node.
func (term BlankNode) String() string {
	return "_:" + term.ID
}

func (term BlankNode) RawValue() string {
	return term.ID
}

// Equal returns whether this blank node is equivalent to another.
func (term BlankNode) Equal(other Term) bool {
	if spec, ok := other.(*BlankNode); ok {
		return term.ID == spec.ID
	}

	return false
}

func term2rdf(t Term) rdf.Term {
	switch t := t.(type) {
	case *BlankNode:
		id := t.RawValue()
		node := rdf.NewBlankNode(id)
		return node
	case *Resource:
		node := rdf.NewIRI(t.RawValue())
		return node
	case *Literal:
		if t.Datatype != nil {
			iri := rdf.NewIRI(t.Datatype.(*Resource).URI)
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
		// id := fmt.Sprint(term.Id)
		return NewBlankNode(term.RawValue())
	case *rdf.Literal:
		if len(term.LanguageTag) > 0 {
			return NewLiteralWithLanguage(term.LexicalForm, term.LanguageTag)
		}
		if term.DatatypeIRI != nil && len(term.DatatypeIRI.String()) > 0 {
			return NewLiteralWithDatatype(term.LexicalForm, NewResource(debrack(term.DatatypeIRI.String())))
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
		// id, _ := strconv.Atoi(term.RawValue())
		return NewBlankNode(term.RawValue())
	case *jsonld.Literal:
		if len(term.Language) > 0 {
			return NewLiteralWithLanguage(term.RawValue(), term.Language)
		}
		if term.Datatype != nil && len(term.Datatype.String()) > 0 {
			return NewLiteralWithDatatype(term.Value, NewResource(term.Datatype.RawValue()))
		}
		return NewLiteral(term.Value)
	case *jsonld.Resource:
		return NewResource(term.RawValue())
	}
	return nil
}

func term2jterm(term Term) jsonld.Term {
	switch term := term.(type) {
	case *BlankNode:
		return jsonld.NewBlankNode(term.RawValue())
	case *Literal:
		if len(term.Language) > 0 {
			return jsonld.NewLiteralWithLanguage(term.Value, term.Language)
		}
		if term.Datatype != nil && len(term.Datatype.String()) > 0 {
			return jsonld.NewLiteralWithDatatype(term.Value, jsonld.NewResource(debrack(term.Datatype.String())))
		}
		return jsonld.NewLiteral(term.Value)
	case *Resource:
		return jsonld.NewResource(term.RawValue())
	}
	return nil
}

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

func atLang(lang string) string {
	if len(lang) > 0 {
		if strings.HasPrefix(lang, "@") {
			return lang
		}
		return "@" + lang
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
