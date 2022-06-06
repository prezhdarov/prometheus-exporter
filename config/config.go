package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	enable = flag.Bool("envflag.enable", false, "Whether to enable reading flags from environment variables additionally to command line. "+
		"Command line flag and file values (if -file is set) have priority over values from environment vars. "+
		"Flags are read only from command line if this flag isn't set.")
	prefix = flag.String("envflag.prefix", "", "Prefix for environment variables if -envflag.enable is set")

	file = flag.String("file", "", "Path to file with configuration data.")
)

func Parse() {

	flag.Parse()

	// Get all flags set on the command line
	flagsSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {

		flagsSet[f.Name] = true

	})

	var fileFlags map[string]string

	//If file set, read the thing
	if *file != "" {

		content, err := ioutil.ReadFile(*file)
		if err != nil {
			log.Fatalf("cannot read file %s - error: %s", *file, err)
		}

		if err = yaml.Unmarshal(content, &fileFlags); err != nil {
			log.Fatalf("cannot read contents of %s - error: %s", *file, err)
		}

		//Now see if any of the flags are already set and if not if there's flags in the file.
		flag.VisitAll(func(f *flag.Flag) {

			if flagsSet[f.Name] {
				return
			}

			if _, exists := fileFlags[f.Name]; exists {

				if err := flag.Set(f.Name, fileFlags[f.Name]); err != nil {
					log.Fatalf("cannot set flag %s to %q, which is read from environment variable %q: %s", f.Name, fileFlags[f.Name], f.Name, err)
					return
				}
				flagsSet[f.Name] = true
			}

		})

	}

	if *enable {

		//Finally for all flags that are not set yet, see if there's corresponding env flag set and get it.
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
