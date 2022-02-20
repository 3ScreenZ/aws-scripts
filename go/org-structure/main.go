package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	structure "github.com/MichaelPalmer1/aws-scripts/go/org-structure/lib"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func main() {
	showIds := flag.Bool("show-ids", false, "Show IDs in the output")
	showAccounts := flag.Bool("show-accounts", false, "Show accounts in the output")
	showPolicies := flag.Bool("show-policies", false, "Show SCPs in the output")
	format := flag.String("format", "text", "Output format {png, json, text}")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg.Region = "us-east-1"
	orgs := organizations.NewFromConfig(cfg)

	roots, err := orgs.ListRoots(context.TODO(), &organizations.ListRootsInput{})
	if err != nil {
		panic(err)
	}

	organization, err := structure.GetChildren(*roots.Roots[0].Id, orgs)
	if err != nil {
		panic(err)
	}

	switch *format {
	case "json":
		bs, err := json.MarshalIndent(organization, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bs))
	case "text":
		printStructure(organization, 0, *showIds, *showAccounts, *showPolicies)
		fmt.Println("\nLegend: \x1B[1mROOT\x1B[0m\t\x1B[4mOU\x1B[0m\t\x1B[3mACCOUNT\x1B[0m")
	case "png":
		buildDiagram(organization, *showIds, *showAccounts, *showPolicies)
	}
}
