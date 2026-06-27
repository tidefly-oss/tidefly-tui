package assets

import _ "embed"

//go:embed docker-compose.yaml
var ComposeProduction []byte

//go:embed docker-compose.dev.yaml
var ComposeDev []byte
