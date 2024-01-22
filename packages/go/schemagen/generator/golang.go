// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"path/filepath"

	"github.com/dave/jennifer/jen"
	"github.com/specterops/bloodhound/schemagen/model"
)

const (
	GraphPackageName  = "github.com/specterops/bloodhound/dawgs/graph"
	SchemaPackageName = "github.com/specterops/bloodhound/graphschema"
	SchemaSourceName  = "github.com/specterops/bloodhound/-/tree/main/packages/cue/schemas"
)

func WriteGolangKindDefinitions(root *jen.File, values []model.StringEnum) {
	root.Var().
		DefsFunc(func(group *jen.Group) {
			for _, value := range values {
				group.Id(value.Symbol).Op("=").Qual(GraphPackageName, "StringKind").Call(jen.Lit(value.GetRepresentation()))
			}
		})
}

func WriteGolangStringEnumeration(root *jen.File, enumTypeSymbol string, values []model.StringEnum) {
	// The enumeration type alias
	root.Type().Id(enumTypeSymbol).String()

	// Generates the string constants for the enumeration type alias
	root.Const().
		DefsFunc(func(group *jen.Group) {
			for _, value := range values {
				group.Id(value.Symbol).Id(enumTypeSymbol).Op("=").Lit(value.GetRepresentation())
			}
		})

	var allString string
	if enumTypeSymbol == "Property" {
		allString = "AllProperties"
	} else {
		allString = "All" + enumTypeSymbol + "s"
	}

	// Generates a function named "All<EnumSymbol>s" that returns a slice of all enumeration instances
	root.Func().Id(allString).
		Params().
		Index().Id(enumTypeSymbol).
		Block(
			jen.Return(jen.Index().Id(enumTypeSymbol).ValuesFunc(func(group *jen.Group) {
				for _, value := range values {
					group.Id(value.Symbol)
				}
			})),
		)

	// Generates a function named "Parse<EnumSymbol" that parses raw string values and returns the typed enumeration
	// instance that matches
	parseFuncSymbol := "Parse" + enumTypeSymbol

	root.Func().Id(parseFuncSymbol).
		Params(jen.Id("source").String()).
		Params(jen.Id(enumTypeSymbol), jen.Error()).
		BlockFunc(func(group *jen.Group) {
			group.Switch(jen.Id("source")).BlockFunc(func(group *jen.Group) {
				for _, value := range values {
					group.Case(jen.Lit(value.GetRepresentation())).Return(jen.Id(value.Symbol), jen.Nil())
				}

				group.Default().Block(
					jen.Return(
						jen.Lit(""),
						jen.Qual("errors", "New").Call(jen.Lit("Invalid enumeration value: ").Op("+").Id("source")),
					),
				)
			})
		})

	// Generates a receiver function named "<EnumSymbol>.String()" that returns the raw string value of the enumeration
	// symbol
	root.Func().Params(jen.Id("s").Id(enumTypeSymbol)).Id("String").
		Params().
		String().
		BlockFunc(func(group *jen.Group) {
			group.Switch(jen.Id("s")).BlockFunc(func(group *jen.Group) {
				for _, value := range values {
					group.Case(jen.Id(value.Symbol)).Return(jen.String().Call(jen.Id(value.Symbol)))
				}

				group.Default().Block(
					jen.Return(
						jen.Lit("Invalid enumeration case: ").Op("+").String().Call(jen.Id("s")),
					),
				)
			})
		})

	// Generates a receiver function named "<EnumSymbol>.Name()" that returns a well-formatted string of the enumeration
	// instance that is more suitable for user interpretation
	root.Func().Params(jen.Id("s").Id(enumTypeSymbol)).Id("Name").
		Params().
		String().
		BlockFunc(func(group *jen.Group) {
			group.Switch(jen.Id("s")).BlockFunc(func(group *jen.Group) {
				for _, value := range values {
					group.Case(jen.Id(value.Symbol)).Return(jen.Lit(value.GetName()))
				}

				group.Default().Block(
					jen.Return(
						jen.Lit("Invalid enumeration case: ").Op("+").String().Call(jen.Id("s")),
					),
				)
			})
		})

	// Generates a receiver function named "<EnumSymbol>.Is(...Kind)" that returns true if any of the variadic arguments
	// contain an enumeration instance that matches this one
	root.Func().Params(jen.Id("s").Id(enumTypeSymbol)).Id("Is").
		Params(jen.Id("others").Op("...").Qual(GraphPackageName, "Kind")).
		Bool().
		BlockFunc(func(group *jen.Group) {
			group.For(jen.List(jen.Id("_"), jen.Id("other")).Op(":=").Range().Id("others")).
				BlockFunc(func(group *jen.Group) {
					group.If(
						jen.Id("value").Op(",").Err().Op(":=").Id(parseFuncSymbol).Call(
							jen.Id("other").Dot("String").Call(),
						),
						jen.Err().Op("==").Nil().Op("&&").Id("value").Op("==").Id("s"),
					).Block(
						jen.Return(jen.Lit(true)),
					)
				})

			group.Return(jen.False())
		})
}

func GenerateGolangAzure(pkgName, dir string, azureSchema model.Azure) error {
	var (
		root  = jen.NewFile(pkgName)
		kinds = append(azureSchema.NodeKinds, azureSchema.RelationshipKinds...)
	)

	root.HeaderComment(fmt.Sprintf("// Code generated by Cuelang code gen. DO NOT EDIT!\n// Cuelang source: %s/", SchemaSourceName))

	WriteGolangKindDefinitions(root, kinds)
	WriteGolangStringEnumeration(root, "Property", azureSchema.Properties)

	root.Func().Id("Relationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationship := range azureSchema.RelationshipKinds {
					group.Id(relationship.Symbol)
				}
			}),
		),
	)

	root.Func().Id("AppRoleTransitRelationshipKinds").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationship := range azureSchema.AppRoleTransitRelationshipKinds {
					group.Id(relationship.Symbol)
				}
			}),
		),
	)

	root.Func().Id("AbusableAppRoleRelationshipKinds").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationship := range azureSchema.AbusableAppRoleRelationshipKinds {
					group.Id(relationship.Symbol)
				}
			}),
		),
	)

	root.Func().Id("ControlRelationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationship := range azureSchema.ControlRelationshipKinds {
					group.Id(relationship.Symbol)
				}
			}),
		),
	)

	root.Func().Id("ExecutionPrivileges").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationship := range azureSchema.ExecutionPrivilegeKinds {
					group.Id(relationship.Symbol)
				}
			}),
		),
	)

	root.Func().Id("PathfindingRelationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, pathRel := range azureSchema.PathfindingRelationships {
					group.Id(pathRel.Symbol)
				}
			}),
		),
	)

	root.Func().Id("NodeKinds").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, nodeKind := range azureSchema.NodeKinds {
					group.Id(nodeKind.Symbol)
				}
			}),
		),
	)

	return WriteSourceFile(root, filepath.Join(dir, "azure.go"))
}

func GenerateGolangSchemaTypes(pkgName, dir string) error {
	root := jen.NewFile(pkgName)

	root.HeaderComment(fmt.Sprintf("// Code generated by Cuelang code gen. DO NOT EDIT!\n// Cuelang source: %s/", SchemaSourceName))

	root.Type().Id("KindDescriptor").StructFunc(func(group *jen.Group) {
		group.Id("Kind").Qual(GraphPackageName, "Kind")
		group.Id("Name").String()
	})

	root.Func().
		Params(jen.Id("s").Id("KindDescriptor")).
		Id("GetName").
		Params().
		String().
		BlockFunc(func(group *jen.Group) {
			group.If(jen.Id("s").Dot("Name").Op("==").Lit("")).BlockFunc(func(group *jen.Group) {
				group.Return(jen.Id("s").Dot("Kind").Dot("String").Call())
			})

			group.Return(jen.Id("s").Dot("Name"))
		})

	root.Type().Id("Path").StructFunc(func(group *jen.Group) {
		group.Id("Outbound").Id("KindDescriptor")
		group.Id("Inbound").Id("KindDescriptor")
		group.Id("Relationships").Index().Id("KindDescriptor")
	})

	return WriteSourceFile(root, filepath.Join(dir, "graph.go"))
}

func GenerateGolangGraphModel(pkgName, dir string, graphSchema model.Graph) error {
	var (
		root  = jen.NewFile(pkgName)
		kinds = append(graphSchema.NodeKinds, graphSchema.RelationshipKinds...)
	)

	root.HeaderComment(fmt.Sprintf("// Code generated by Cuelang code gen. DO NOT EDIT!\n// Cuelang source: %s/", SchemaSourceName))

	WriteGolangKindDefinitions(root, kinds)

	if len(graphSchema.Properties) > 0 {
		WriteGolangStringEnumeration(root, "Property", graphSchema.Properties)
	}

	root.Func().Id("Nodes").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, nodeKind := range graphSchema.NodeKinds {
					group.Id(nodeKind.Symbol)
				}
			}),
		),
	)

	root.Func().Id("Relationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationshipKind := range graphSchema.RelationshipKinds {
					group.Id(relationshipKind.Symbol)
				}
			}),
		),
	)

	root.Func().Id("NodeKinds").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, nodeKind := range graphSchema.NodeKinds {
					group.Id(nodeKind.Symbol)
				}
			}),
		),
	)

	return WriteSourceFile(root, filepath.Join(dir, pkgName+".go"))
}

func GenerateGolangActiveDirectory(pkgName, dir string, adSchema model.ActiveDirectory) error {
	var (
		root  = jen.NewFile(pkgName)
		kinds = append(adSchema.NodeKinds, adSchema.RelationshipKinds...)
	)

	root.HeaderComment(fmt.Sprintf("// Code generated by Cuelang code gen. DO NOT EDIT!\n// Cuelang source: %s/", SchemaSourceName))

	WriteGolangKindDefinitions(root, kinds)
	WriteGolangStringEnumeration(root, "Property", adSchema.Properties)

	root.Func().Id("Nodes").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, nodeKind := range adSchema.NodeKinds {
					group.Id(nodeKind.Symbol)
				}
			}),
		),
	)

	root.Func().Id("Relationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, relationshipKind := range adSchema.RelationshipKinds {
					group.Id(relationshipKind.Symbol)
				}
			}),
		),
	)

	root.Func().Id("ACLRelationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, aclRelationship := range adSchema.ACLRelationships {
					group.Id(aclRelationship.Symbol)
				}
			}),
		),
	)

	root.Func().Id("PathfindingRelationships").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, pathRelationship := range adSchema.PathfindingRelationships {
					group.Id(pathRelationship.Symbol)
				}
			}),
		),
	)

	root.Func().
		Id("IsACLKind").
		Params(jen.Id("s").Qual(GraphPackageName, "Kind")).
		Bool().
		BlockFunc(func(group *jen.Group) {
			group.For(jen.Id("_").Op(",").Id("acl").Op(":=").Range().Id("ACLRelationships").Call()).Block(
				jen.If(
					jen.Id("s").Op("==").Id("acl"),
				).Block(
					jen.Return(jen.Lit(true)),
				))

			group.Return(jen.Lit(false))
		})

	root.Func().Id("NodeKinds").Params().Index().Qual(GraphPackageName, "Kind").Block(
		jen.Return(
			jen.Index().Qual(GraphPackageName, "Kind").ValuesFunc(func(group *jen.Group) {
				for _, nodeKind := range adSchema.NodeKinds {
					group.Id(nodeKind.Symbol)
				}
			}),
		),
	)

	return WriteSourceFile(root, filepath.Join(dir, "ad.go"))
}
