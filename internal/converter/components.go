package converter

import (
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func fileToComponents(fd protoreflect.FileDescriptor) (openapi31.Components, error) {
	// Add schema from messages/enums
	components := openapi31.Components{}
	rootSchema := fileToSchema(&State{}, fd)
	for _, item := range rootSchema.Items.SchemaArray {
		if item.TypeObject == nil {
			continue
		}
		m, err := item.ToSimpleMap()
		if err != nil {
			return components, err
		}
		components.WithSchemasItem(*item.TypeObject.ID, m)
	}

	// Add requestBodies and responses for methods
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < services.Len(); j++ {
			method := methods.Get(j)
			op := &openapi31.Operation{}
			op.WithTags(string(service.FullName()))
			loc := fd.SourceLocations().ByDescriptor(method)
			op.WithDescription(formatComments(loc))

			// Request Body
			trueVar := true
			components.WithRequestBodiesItem(formatTypeRef(string(method.Input().FullName())),
				openapi31.RequestBodyOrReference{
					RequestBody: &openapi31.RequestBody{
						Description: new(string),
						Content: map[string]openapi31.MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"$ref": "#/components/schemas/" + formatTypeRef(string(method.Input().FullName())),
								},
							},
						},
						Required:      &trueVar,
						MapOfAnything: map[string]interface{}{},
					},
				},
			)

			components.WithResponsesItem(formatTypeRef(string(method.Output().FullName())),
				openapi31.ResponseOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/schemas/" + formatTypeRef(string(method.Output().FullName())),
					},
				},
			)
		}
	}

	// Add our own type for errors
	reflector := jsonschema.Reflector{}
	connectError, err := reflector.Reflect(ConnectError{})
	if err != nil {
		return components, err
	}
	connectError.WithID("connect.error")

	components.WithSchemasItem(*connectError.ID, map[string]interface{}{
		"$id":         connectError.ID,
		"description": connectError.Description,
		"properties":  connectError.Properties,
		"title":       connectError.Title,
		"type":        connectError.Type,
	})

	components.WithResponsesItem("connect.error", openapi31.ResponseOrReference{
		Reference: &openapi31.Reference{
			Ref: "#/components/schemas/connect.error",
		},
	})

	return components, nil
}

type ConnectError struct {
	Code    string `json:"code" example:"CodeNotFound" enum:"CodeCanceled,CodeUnknown,CodeInvalidArgument,CodeDeadlineExceeded,CodeNotFound,CodeAlreadyExists,CodePermissionDenied,CodeResourceExhausted,CodeFailedPrecondition,CodeAborted,CodeOutOfRange,CodeInternal,CodeUnavailable,CodeDataLoss,CodeUnauthenticated"`
	Message string `json:"message,omitempty"`
}