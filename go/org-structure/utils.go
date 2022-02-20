package main

import (
	"fmt"
	"strings"

	structure "github.com/MichaelPalmer1/aws-scripts/go/org-structure/lib"
	"github.com/blushft/go-diagrams/diagram"
	awsNodes "github.com/blushft/go-diagrams/nodes/aws"
)

func buildDiagram(orgStructure *structure.Structure, showIds, showAccounts, showPolicies bool) {
	d, err := diagram.New(diagram.Filename("organization"), diagram.Direction("TB"))
	if err != nil {
		panic(err)
	}

	var renderChild func(*structure.Structure, *diagram.Node)
	renderChild = func(childStructure *structure.Structure, parentNode *diagram.Node) {
		var node *diagram.Node

		// Build node
		text := childStructure.Name
		if showIds {
			text += fmt.Sprintf("\n(%s)", childStructure.Id)
		}
		if showPolicies {
			text += fmt.Sprintf("\n[%s]", strings.Join(childStructure.Policies, ", "))
		}
		nodeLabel := diagram.NodeLabel(text)
		node = awsNodes.Management.Organizations(nodeLabel)

		if parentNode != nil {
			d.Connect(parentNode, node)
		}

		// Render child org units
		for _, orgUnit := range childStructure.OrgUnits {
			renderChild(&orgUnit, node)
		}

		// Render child accounts
		if showAccounts {
			for _, account := range childStructure.Accounts {
				renderChild(&account, node)
			}
		}
	}

	renderChild(orgStructure, nil)

	if err := d.Render(); err != nil {
		panic(err)
	}
}

func printStructure(orgStructure *structure.Structure, depth int, showIds bool, showAccounts bool, showPolicies bool) {
	text := ""
	switch orgStructure.Type {
	case "ROOT":
		text += "\x1B[1m"
	case "ORGANIZATION_UNIT":
		text += "\x1B[4m"
	case "ACCOUNT":
		text += "\x1B[3m"
	}

	text += orgStructure.Name
	text += "\x1B[0m"

	if showIds {
		text += fmt.Sprintf(" (%s)", orgStructure.Id)
	}
	if showPolicies {
		text += fmt.Sprintf("\t[%s]", strings.Join(orgStructure.Policies, ", "))
	}

	if depth == 0 && orgStructure.Type == "ROOT" {
		fmt.Println(text)
	} else if depth == 1 {
		fmt.Println("|-- " + text)
	} else {
		for i := 0; i < depth; i++ {
			fmt.Print("  ")
		}
		fmt.Println("|-- " + text)
	}

	for _, orgUnit := range orgStructure.OrgUnits {
		printStructure(&orgUnit, depth+1, showIds, showAccounts, showPolicies)
	}

	if showAccounts {
		for _, account := range orgStructure.Accounts {
			printStructure(&account, depth+1, showIds, showAccounts, showPolicies)
		}
	}
}
