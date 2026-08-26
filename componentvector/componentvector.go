// SPDX-FileCopyrightText: Contributors to the Gardener project
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
