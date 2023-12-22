// SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logranger

import "errors"

// ErrCertConfigEmpty is returned if a TLS listener is configured but ot certificate
// or key paths are set
var ErrCertConfigEmpty = errors.New("certificate and key paths are required for listener type: TLS")
