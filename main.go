/*
Copyright © 2022 roncewind <dad@lynntribe.net>
*/
package main

import (
	"log"

	"github.com/roncewind/load/cmd"
)

func main() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Lmicroseconds | log.LUTC)
	cmd.Execute()
}
