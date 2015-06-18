// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"log"
	"os"

	"code.google.com/p/gcfg"
)

type FactoidConfig struct {
	Wallet struct {
		Address          string
		Port             int
		DataFile         string
		RefreshInSeconds string
	}
}

const defaultConfig = `
; ------------------------------------------------------------------------------
; App settings
; ------------------------------------------------------------------------------
[wallet]
Address = localhost
Port = 8089
DataFile = /tmp/fctwallet.dat
RefreshInSeconds = 60
`

// ReadConfig reads the factoid.conf file and returns the FactoidConfig object
func ReadConfig() *FactoidConfig {
	cfg := new(FactoidConfig)
	filename := os.Getenv("HOME") + "/.factom/factoid.conf"
	if err := gcfg.ReadFileInto(cfg, filename); err != nil {
		gcfg.ReadStringInto(cfg, defaultConfig)
	}
	return cfg
}
