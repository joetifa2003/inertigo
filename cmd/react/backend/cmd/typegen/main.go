package main

import (
	"log"
	"os"

	"github.com/joetifa2003/inertigo/typegen"
)

func main() {
	out, err := typegen.GenerateTypes(
		typegen.Package{
			Path:   "github.com/joetifa2003/inertigo/cmd/react/backend/models/pages",
			Prefix: "Page",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("dist", 0755)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("dist/index.ts")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(out)
	if err != nil {
		log.Fatal(err)
	}
}
