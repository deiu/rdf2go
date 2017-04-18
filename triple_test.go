package rdf2go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var one = NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))

func TestTripleEquals(t *testing.T) {
	assert.True(t, one.Equal(NewTriple(NewResource("a"), NewResource("b"), NewResource("c"))))
}

func TestTripleString(t *testing.T) {
	assert.Equal(t, "<a> <b> <c> .", one.String())
}
