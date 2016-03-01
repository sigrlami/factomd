// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	. "github.com/FactomProject/factomd/database/blockExtractor"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"os"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("BlockExtractor level/bolt [ChainID-To-Extract]")
	fmt.Println("Leave out the last one to export basic chains (A, D, EC, F)")
	if len(os.Args) < 1 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	levelBolt := os.Args[1]

	if levelBolt != level && levelBolt != bolt {
		fmt.Println("\nFirst argument should be `level` or `bolt`")
		os.Exit(1)
	}

	chainID := ""
	if len(os.Args) == 3 {
		chainID = os.Args[2]
	}

	state := new(state.State)
	state.Cfg = util.ReadConfig("")
	if levelBolt == level {
		err := state.InitLevelDB()
		if err != nil {
			panic(err)
		}
	}
	if levelBolt == bolt {
		err := state.InitBoltDB()
		if err != nil {
			panic(err)
		}
	}
	dbo := state.GetDB()

	if chainID != "" {
		err := ExportEChain(chainID, dbo)
		if err != nil {
			panic(err)
		}
	} else {
		err := ExportDChain(dbo)
		if err != nil {
			panic(err)
		}
		err = ExportECChain(dbo)
		if err != nil {
			panic(err)
		}
		err = ExportAChain(dbo)
		if err != nil {
			panic(err)
		}
		err = ExportFctChain(dbo)
		if err != nil {
			panic(err)
		}
		err = ExportDirBlockInfo(dbo)
		if err != nil {
			panic(err)
		}
	}
}