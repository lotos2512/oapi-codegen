package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A property written as allOf [$ref, {description}] keeps the referenced
// type: no inline duplicate of the component, no format merge error.
const annotationOnlySpec = `
openapi: 3.0.1
info: {title: t, version: "1"}
paths:
  /items:
    get:
      operationId: listItems
      parameters:
        - $ref: "#/components/parameters/NextCursor"
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Item"}
components:
  parameters:
    NextCursor:
      name: cursor
      in: query
      schema:
        $ref: "#/components/schemas/NextCursor"
  schemas:
    NextCursor:
      type: string
      description: cursor of the next page
    ItemStatus:
      type: string
      enum: [active, inactive]
    ArticleNumber:
      type: integer
      format: int32
      description: article number in the cart
    Item:
      type: object
      properties:
        status:
          allOf:
            - $ref: "#/components/schemas/ItemStatus"
            - description: status of this item
        article_number:
          allOf:
            - $ref: "#/components/schemas/ArticleNumber"
            - description: article number in the refund cart
        next_cursor:
          $ref: "#/components/schemas/NextCursor"
`

func TestGenerate_AllOfWithAnnotationOnlyMember(t *testing.T) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(annotationOnlySpec))
	require.NoError(t, err)

	code, err := Generate(swagger, Configuration{
		PackageName: "testpkg",
		Generate:    GenerateOptions{Client: true, Models: true},
	})
	require.NoError(t, err)

	assert.Regexp(t, `Status\s+\*ItemStatus\s`, code)
	assert.Regexp(t, `ArticleNumber\s+\*ArticleNumber\s`, code)
	assert.Regexp(t, `Cursor\s+\*NextCursor\s`, code)
}

func TestWithoutAnnotationOnly_KeepsAtLeastOne(t *testing.T) {
	only := []*openapi3.SchemaRef{{Value: &openapi3.Schema{Description: "d"}}}
	assert.Len(t, withoutAnnotationOnly(only), 1)

	typed := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Description: "d"}}
	ref := &openapi3.SchemaRef{Ref: "#/components/schemas/X", Value: &openapi3.Schema{}}
	assert.Equal(t, []*openapi3.SchemaRef{ref, typed}, withoutAnnotationOnly([]*openapi3.SchemaRef{ref, typed}))
}

func TestTypeDefinitionsEquivalent_IgnoresAnnotations(t *testing.T) {
	a := TypeDefinition{TypeName: "NextCursor", Schema: Schema{GoType: "string",
		OAPISchema: &openapi3.Schema{Type: &openapi3.Types{"string"}, Description: "one"}}}
	b := TypeDefinition{TypeName: "NextCursor", Schema: Schema{GoType: "string",
		OAPISchema: &openapi3.Schema{Type: &openapi3.Types{"string"}, Description: "two", Example: "x"}}}
	assert.True(t, TypeDefinitionsEquivalent(a, b))

	c := TypeDefinition{TypeName: "NextCursor", Schema: Schema{GoType: "int",
		OAPISchema: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}}
	assert.False(t, TypeDefinitionsEquivalent(a, c))
}
