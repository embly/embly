package qm

import (
	"testing"

	"github.com/embly/vinyl/vinyl-go/transport"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	qm := And(
		Field("price").LessThan(50),
		Field("flower").Matches(Field("type").Equals("ROSE")),
	)
	qc, errs := qm.QueryComponent()
	if len(errs) > 0 {
		t.Error(errs)
	}
	assert.Equal(t, transport.QueryComponent_AND, qc.ComponentType)
	assert.Equal(t, "price", qc.Children[0].Field.Name)
	assert.Equal(t, transport.Field_LESS_THAN, qc.Children[0].Field.ComponentType)
	assert.Equal(t, int64(50), qc.Children[0].Field.Value.Int64)
	assert.Equal(t, "flower", qc.Children[1].Field.Name)
	assert.Equal(t, transport.Field_MATCHES, qc.Children[1].Field.ComponentType)
	assert.Equal(t, "type", qc.Children[1].Field.Matches.Field.Name)
	assert.Equal(t, transport.Field_EQUALS, qc.Children[1].Field.Matches.Field.ComponentType)
	assert.Equal(t, "ROSE", qc.Children[1].Field.Matches.Field.Value.String_)
}

func TestError(t *testing.T) {
	qm := And(
		And(
			Field("price").LessThan(50),
			Field("flower").Matches(Field("type").Equals("ROSE")),
		),
		And(
			Field("price").LessThan(50),
			Field("flower").Matches(Field("type").Equals("ROSE")),
		),
	)
	_ = qm

}
