package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	enable = flag.Bool("envflag.enable", false, "Whether to enable reading flags from environment variables additionally to command line. "+
		"Command line flag values have priority over values from environment vars. "+
		"Flags are read only from command line if this flag isn't set.")
	prefix = flag.String("envflag.prefix", "", "Prefix for environment variables if -envflag.enable is set")
)

func Parse() {
	flag.Parse()
	if !*enable {
		return
	}

	flagsSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {

		flagsSet[f.Name] = true

	})

	flag.VisitAll(func(f *flag.Flag) {

		if flagsSet[f.Name] {

			return

		}
		fname := getEnvFlagName(f.Name)
		if v, ok := os.LookupEnv(fname); ok {

			if err := flag.Set(f.Name, v); err != nil {

				log.Fatalf("cannot set flag %s to %q, which is read from environment variable %q: %s", f.Name, v, fname, err)

			}

		}
	})
}

func getEnvFlagName(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	return *prefix + s
}

func Usage(s string) {
	f := flag.CommandLine.Output()
	fmt.Fprintf(f, "%s\n", s)
	if hasHelpFlag(os.Args[1:]) {
		flag.PrintDefaults()
	} else {
		fmt.Fprintf(f, `Run "%s -help" in order to see the description for all the available flags`+"\n", os.Args[0])
	}
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if isHelpArg(arg) {
			return true
		}
	}
	return false
}

func isHelpArg(arg string) bool {
	if !strings.HasPrefix(arg, "-") {
		return false
	}
	arg = strings.TrimPrefix(arg[1:], "-")
	return arg == "h" || arg == "help"
}
