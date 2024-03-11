package grpc

// a composite source to support protos and reflection options
// mainly a copy paste from grpcurl
// https://github.com/fullstorydev/grpcurl/blob/70c215f7e2c272fdccb9386bbbff5dbc49fdd4fb/cmd/grpcurl/grpcurl.go#L236

import (
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
)

type compositeSource struct {
	reflection grpcurl.DescriptorSource
	file       grpcurl.DescriptorSource
}

func (cs compositeSource) ListServices() ([]string, error) {
	if svcs, err := cs.file.ListServices(); err == nil {
		return svcs, err
	}

	return cs.reflection.ListServices()
}

func (cs compositeSource) FindSymbol(fullyQualifiedName string) (desc.Descriptor, error) {
	d, err := cs.file.FindSymbol(fullyQualifiedName)
	if err == nil {
		return d, nil
	}
	return cs.reflection.FindSymbol(fullyQualifiedName)
}

func (cs compositeSource) AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error) {
	exts, err := cs.reflection.AllExtensionsForType(typeName)
	if err != nil {
		// On error fall back to file source
		return cs.file.AllExtensionsForType(typeName)
	}
	// Track the tag numbers from the reflection source
	tags := make(map[int32]bool)
	for _, ext := range exts {
		tags[ext.GetNumber()] = true
	}
	fileExts, err := cs.file.AllExtensionsForType(typeName)
	if err != nil {
		return exts, nil
	}
	for _, ext := range fileExts {
		// Prioritize extensions found via reflection
		if !tags[ext.GetNumber()] {
			exts = append(exts, ext)
		}
	}
	return exts, nil
}
