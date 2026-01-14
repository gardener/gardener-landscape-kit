// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:generate ../hack/generate-componentname-constants.sh

package componentvector

import (
	_ "embed"
)

var (
	//go:embed components.yaml
	DefaultComponentsYAML []byte
)
