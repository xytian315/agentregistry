package v1alpha1

import "fmt"

// Validate runs structural validation on the MCPServer envelope.
func (m *MCPServer) Validate() error {
	var errs FieldErrors
	errs = append(errs, ValidateObjectMeta(m.Metadata)...)
	errs = append(errs, validateMCPServerSpec(&m.Spec)...)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func validateMCPServerSpec(s *MCPServerSpec) FieldErrors {
	var errs FieldErrors
	errs.Append("spec.title", validateTitle(s.Title))
	if s.Source != nil {
		for _, e := range validateRepository(s.Source.Repository) {
			errs.Append("spec.source."+e.Path, e.Cause)
		}
		if pkg := s.Source.Package; pkg != nil {
			if pkg.RegistryType == "" {
				errs.Append("spec.source.package.registryType", fmt.Errorf("%w", ErrRequiredField))
			}
			if pkg.Identifier == "" {
				errs.Append("spec.source.package.identifier", fmt.Errorf("%w", ErrRequiredField))
			}
			if pkg.Transport.Type == "" {
				errs.Append("spec.source.package.transport.type", fmt.Errorf("%w", ErrRequiredField))
			}
		}
	}

	return errs
}
