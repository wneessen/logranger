package logranger

import "errors"

// ErrCertConfigEmpty is returned if a TLS listener is configured but not certificate or key paths are set
var ErrCertConfigEmpty = errors.New("certificate and key paths are required for listener type: TLS")
