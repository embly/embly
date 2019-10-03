// Package qm provides query building funcitonality for vinyl-go.
//     db.First(And(
//        Field("price").LessThan(50),
//        Field("flower").Matches(Field("type").Equals("ROSE")),
//     ))
//
package qm

// And(Field("price").LessThan(50), Field("flower").Matches(Field("type").Equals("ROSE")))
// First(msg, And(Field("email").EqualTo("max@max.com"), Field)
import (
	"errors"

	"github.com/embly/vinyl/vinyl-go/transport"
)

type queryComponentMeta struct {
	queryComponent *transport.QueryComponent
	errors         []error
}

// And check that a set of components all evaluate to true for a given record.
func And(qcs ...QueryComponent) QueryComponent {
	qcm := &queryComponentMeta{queryComponent: &transport.QueryComponent{}}
	qcm.queryComponent.Children = make([]*transport.QueryComponent, len(qcs))
	for i, qc := range qcs {
		qcptr, errs := qc.QueryComponent()
		qcm.queryComponent.Children[i] = qcptr
		qcm.errors = append(qcm.errors, errs...)
	}
	qcm.queryComponent.ComponentType = transport.QueryComponent_AND
	return qcm
}

// Or check that any of a set of components evaluate to true for a given record.
func Or(qcs ...QueryComponent) QueryComponent {
	qcm := &queryComponentMeta{queryComponent: &transport.QueryComponent{}}
	qcm.queryComponent.Children = make([]*transport.QueryComponent, len(qcs))
	for i, qc := range qcs {
		qcptr, errs := qc.QueryComponent()
		qcm.queryComponent.Children[i] = qcptr
		qcm.errors = append(qcm.errors, errs...)
	}
	qcm.queryComponent.ComponentType = transport.QueryComponent_OR
	return qcm
}

// Not negate a component test
func Not(qc QueryComponent) QueryComponent {
	qcm := &queryComponentMeta{}
	var errs []error
	qcm.queryComponent.Child, errs = qc.QueryComponent()
	qcm.errors = append(qcm.errors, errs...)
	qcm.queryComponent.ComponentType = transport.QueryComponent_NOT
	return qcm
}

func (qcm *queryComponentMeta) QueryComponent() (*transport.QueryComponent, []error) {
	return qcm.queryComponent, qcm.errors
}

// QueryComponent interfaces can return query component metadata and errors
type QueryComponent interface {
	QueryComponent() (*transport.QueryComponent, []error)
}

// FieldMeta holds the metadata for asserting about a field value.
type FieldMeta struct {
	field  *transport.Field
	errors []error
}

// QueryComponent return the query component and errors for this field
func (fm *FieldMeta) QueryComponent() (*transport.QueryComponent, []error) {
	return &transport.QueryComponent{
		ComponentType: transport.QueryComponent_FIELD,
		Field:         fm.field,
	}, fm.errors
}

// Field context for asserting about a field value.
func Field(name string) *FieldMeta {
	return &FieldMeta{
		field: &transport.Field{
			Name: name,
		},
	}
}

// ValueForInterface translates any valid type into a protobuf value
func ValueForInterface(value interface{}) (out *transport.Value, err error) {
	switch value := value.(type) {
	case float64:
		out = &transport.Value{
			ValueType: transport.Value_DOUBLE,
			Double:    value,
		}
	case float32:
		out = &transport.Value{
			ValueType: transport.Value_FLOAT,
			Float:     value,
		}
	case int32:
		out = &transport.Value{
			ValueType: transport.Value_INT32,
			Int32:     value,
		}
	case int64:
		out = &transport.Value{
			ValueType: transport.Value_INT64,
			Int64:     value,
		}
	case int:
		// TODO: is this ok?
		out = &transport.Value{
			ValueType: transport.Value_INT64,
			Int64:     int64(value),
		}
	case bool:
		out = &transport.Value{
			ValueType: transport.Value_BOOL,
			Bool:      value,
		}
	case string:
		out = &transport.Value{
			ValueType: transport.Value_STRING,
			String_:   value,
		}
	case []byte:
		out = &transport.Value{
			ValueType: transport.Value_BYTES,
			Bytes:     value,
		}
	default:
		err = errors.New("type not supported")
	}
	return
}

func (fm *FieldMeta) valueForInterface(value interface{}) *transport.Value {
	out, err := ValueForInterface(value)
	if err != nil {
		fm.errors = append(fm.errors, err)
		out = &transport.Value{}
	}
	return out
}

// Equals used like `Field("foo").Equals("bar")`
func (fm *FieldMeta) Equals(value interface{}) QueryComponent {
	fm.field.ComponentType = transport.Field_EQUALS
	fm.field.Value = fm.valueForInterface(value)
	return fm
}

// GreaterThan checks if the field has a value greater than the given comparand
func (fm *FieldMeta) GreaterThan(value interface{}) QueryComponent {
	fm.field.ComponentType = transport.Field_GREATER_THAN
	fm.field.Value = fm.valueForInterface(value)
	return fm
}

// LessThan checks if the field has a value less than the given comparand.
func (fm *FieldMeta) LessThan(value interface{}) QueryComponent {
	fm.field.ComponentType = transport.Field_LESS_THAN
	fm.field.Value = fm.valueForInterface(value)
	return fm
}

// IsEmpty returns true if the repeated field does not have any occurrences.
func (fm *FieldMeta) IsEmpty() QueryComponent {
	fm.field.ComponentType = transport.Field_EMPTY
	return fm
}

// NotEmpty returns true if the repeated field has occurrences.
func (fm *FieldMeta) NotEmpty() QueryComponent {
	fm.field.ComponentType = transport.Field_NOT_EMPTY
	return fm
}

// IsNull returns true if the field has not been set and uses the default value.
func (fm *FieldMeta) IsNull() QueryComponent {
	fm.field.ComponentType = transport.Field_IS_NULL
	return fm
}

// Matches ..
func (fm *FieldMeta) Matches(qc QueryComponent) QueryComponent {
	fm.field.ComponentType = transport.Field_MATCHES
	var errs []error
	fm.field.Matches, errs = qc.QueryComponent()
	fm.errors = append(fm.errors, errs...)
	return fm
}
