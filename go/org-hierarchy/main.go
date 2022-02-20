package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

type Child struct {
	Id   string
	Name string
	Type string
}

var orgs *organizations.Client
var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(`\d{12}`)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg.Region = "us-east-1"
	orgs = organizations.NewFromConfig(cfg)
}

func GetHierarchy(childId string) ([]Child, error) {
	var hierarchy []Child

	if strings.HasPrefix(childId, "r-") {
		return []Child{
			{
				Id:   childId,
				Name: "Root",
				Type: "ROOT",
			},
		}, nil
	} else if strings.HasPrefix(childId, "ou-") {
		ouOutput, err := orgs.DescribeOrganizationalUnit(context.TODO(), &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: aws.String(childId),
		})
		if err != nil {
			return nil, err
		}

		hierarchy = append(hierarchy, Child{
			Id:   childId,
			Name: *ouOutput.OrganizationalUnit.Name,
			Type: "ORGANIZATIONAL_UNIT",
		})
	} else if re.MatchString(childId) {
		acctOutput, err := orgs.DescribeAccount(context.TODO(), &organizations.DescribeAccountInput{
			AccountId: aws.String(childId),
		})
		if err != nil {
			return nil, err
		}

		hierarchy = append(hierarchy, Child{
			Id:   childId,
			Name: *acctOutput.Account.Name,
			Type: "ACCOUNT",
		})
	} else {
		return nil, errors.New("Unknown child id format " + childId)
	}

	parentOutput, err := orgs.ListParents(context.TODO(), &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	})
	if err != nil {
		return nil, err
	}

	childHierarchy, err := GetHierarchy(*parentOutput.Parents[0].Id)
	if err != nil {
		return nil, err
	}
	hierarchy = append(hierarchy, childHierarchy...)

	reverse(hierarchy)

	return hierarchy, nil
}

func reverse(input []Child) {
	i := 0
	j := len(input) - 1

	for i < j {
		input[i], input[j] = input[j], input[i]
		i++
		j--
	}
}

func main() {
	account := flag.String("account", "", "AWS Account ID or Name")
	outputFormat := flag.String("format", "text", "Output format {text, json}")
	showIds := flag.Bool("show-ids", false, "Whether to include OU IDs")
	flag.Parse()

	// Get account id from organization
	var accountId string = *account
	if re.MatchString(*account) {
		accountId = *account
	} else {
		// accountId = GetAccountId(account)
	}

	// Get hierarchy
	hierarchy, err := GetHierarchy(accountId)
	if err != nil {
		panic(err)
	}

	// Output formats
	if *outputFormat == "text" {
		var results []string
		if *showIds {
			for _, item := range hierarchy {
				results = append(results, fmt.Sprintf("%s (%s)", item.Name, item.Id))
			}
			fmt.Println(strings.Join(results, " -> "))
		} else {
			for _, item := range hierarchy {
				results = append(results, item.Name)
			}
			fmt.Println(strings.Join(results, " -> "))
		}
	} else if *outputFormat == "json" {
		bs, err := json.MarshalIndent(hierarchy, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bs))
	}

}
