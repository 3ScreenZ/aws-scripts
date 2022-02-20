package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	orgAccount "github.com/MichaelPalmer1/aws-scripts/go/org-account-id/lib"
	hierarchy "github.com/MichaelPalmer1/aws-scripts/go/org-hierarchy/lib"
	orgUtils "github.com/MichaelPalmer1/aws-scripts/go/org-utils"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func main() {
	var account, format string
	var showIds bool
	flag.StringVar(&account, "account", "", "AWS Account ID or Name")
	flag.StringVar(&format, "format", "text", "Output format {text, json}")
	flag.BoolVar(&showIds, "show-ids", false, "Whether to include OU IDs")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg.Region = "us-east-1"
	orgs := organizations.NewFromConfig(cfg)

	// Get account id from organization
	var accountId string = account
	if !orgUtils.AccountRegex.MatchString(account) {
		accountId = *orgAccount.GetAccountId(account, orgs)
	}

	// Get hierarchy
	organization, err := hierarchy.GetHierarchy(accountId, orgs)
	if err != nil {
		panic(err)
	}

	// Output formats
	if format == "text" {
		var results []string
		if showIds {
			for _, item := range organization {
				results = append(results, fmt.Sprintf("%s (%s)", item.Name, item.Id))
			}
			fmt.Println(strings.Join(results, " -> "))
		} else {
			for _, item := range organization {
				results = append(results, item.Name)
			}
			fmt.Println(strings.Join(results, " -> "))
		}
	} else if format == "json" {
		bs, err := json.MarshalIndent(organization, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bs))
	}

}
