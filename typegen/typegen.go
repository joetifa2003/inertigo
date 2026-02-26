package typegen

import (
	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/bindings/walk"
	"github.com/coder/guts/config"
)

type Package struct {
	Path   string
	Prefix string
}

func GenerateTypes(pkgs ...Package) (string, error) {
	golang, err := guts.NewGolangParser()
	if err != nil {
		return "", err
	}

	golang.IncludeCustomDeclaration(config.StandardMappings())
	golang.IncludeGenerateWithPrefix("github.com/joetifa2003/inertigo/props", "Props")
	for _, pkg := range pkgs {
		golang.IncludeGenerateWithPrefix(pkg.Path, pkg.Prefix)
	}

	golang.PreserveComments()

	ts, err := golang.ToTypescript()
	if err != nil {
		return "", err
	}

	ts.ApplyMutations(
		config.ExportTypes,
		unwrapGeneric("PropsDeferred"),
	)

	output, err := ts.Serialize()
	if err != nil {
		return "", err
	}

	return output, nil
}

// UnwrapGeneric returns a mutation that replaces references to the given
// generic type names with their first type argument.
// Names should match the Ref() of the identifier (i.e. prefix + name).
// Example: UnwrapGeneric("UtilsDeferred") turns UtilsDeferred<number> into number.
func unwrapGeneric(names ...string) guts.MutationFunc {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return func(ts *guts.Typescript) {
		ts.ForEach(func(key string, node bindings.Node) {
			walk.Walk(&unwrapVisitor{targets: set}, node)
		})
	}
}

type unwrapVisitor struct {
	targets map[string]struct{}
}

func (v *unwrapVisitor) Visit(node bindings.Node) walk.Visitor {
	switch n := node.(type) {
	case *bindings.PropertySignature:
		if ref, ok := n.Type.(*bindings.ReferenceType); ok {
			if _, match := v.targets[ref.Name.Ref()]; match && len(ref.Arguments) == 1 {
				n.Type = bindings.Union(
					ref.Arguments[0],
					ptr(bindings.KeywordUndefined),
				)
			}
		}
	}
	return v
}

func ptr[T any](v T) *T {
	return &v
}
