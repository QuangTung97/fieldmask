package fieldmask

import (
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser_Simple_Message(t *testing.T) {
	fields := parseMessage(&pb.Provider{})
	assert.Equal(t, []objectField{
		{
			name:     "Id",
			jsonName: "id",
		},
		{
			name:     "Name",
			jsonName: "name",
		},
		{
			name:     "Logo",
			jsonName: "logo",
		},
		{
			name:     "ImageUrl",
			jsonName: "imageUrl",
		},
	}, fields)
}

func TestParser_Complex_Object(t *testing.T) {
	fields := parseMessage(&pb.Product{})
	assert.Equal(t, []objectField{
		{
			name:     "Sku",
			jsonName: "sku",
			subType:  subFieldTypeSimple,
		},
		{
			name:     "Provider",
			jsonName: "provider",
			subFields: []objectField{
				{
					name:     "Id",
					jsonName: "id",
				},
				{
					name:     "Name",
					jsonName: "name",
				},
				{
					name:     "Logo",
					jsonName: "logo",
				},
				{
					name:     "ImageUrl",
					jsonName: "imageUrl",
				},
			},
			subType: subFieldTypeObject,
		},
		{
			name:     "Attributes",
			jsonName: "attributes",
			subType:  subFieldTypeArray,
			subFields: []objectField{
				{
					name:     "Id",
					jsonName: "id",
				},
				{
					name:     "Code",
					jsonName: "code",
				},
				{
					name:     "Name",
					jsonName: "name",
				},
				{
					name:     "Options",
					jsonName: "options",
					subFields: []objectField{
						{
							name:     "Code",
							jsonName: "code",
						},
						{
							name:     "Name",
							jsonName: "name",
						},
					},
					subType: subFieldTypeArray,
				},
			},
		},
	}, fields)
}
