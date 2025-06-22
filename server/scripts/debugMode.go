package scripts

import "flag"

var DebugMode *bool

func init() {
	DebugMode = flag.Bool("debug", false, "")
}
