package structure

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type Structure struct {
	Id       string
	Name     string
	Type     string
	Policies []string
	OrgUnits []Structure
	Accounts []Structure
}

var accountRegexp *regexp.Regexp

func init() {
	accountRegexp = regexp.MustCompile(`\d{12}`)
}

func GetPolicies(parentId string, client *organizations.Client) ([]string, error) {
	var policies []string

	paginator := organizations.NewListPoliciesForTargetPaginator(client, &organizations.ListPoliciesForTargetInput{
		TargetId: aws.String(parentId),
		Filter:   types.PolicyTypeServiceControlPolicy,
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, policy := range output.Policies {
			if !contains(policies, *policy.Name) {
				policies = append(policies, *policy.Name)
			}
		}
	}

	return policies, nil
}

func GetChildren(parentId string, client *organizations.Client) (*Structure, error) {
	// Get SCPs
	policies, err := GetPolicies(parentId, client)
	if err != nil {
		return nil, err
	}

	// Initialize organization structure
	organization := &Structure{
		Id:       parentId,
		Policies: policies,
		OrgUnits: []Structure{},
		Accounts: []Structure{},
	}

	// Configure parent based on id type
	if strings.HasPrefix(parentId, "r-") {
		// Root node
		organization.Name = "Root"
		organization.Type = "ROOT"
	} else if strings.HasPrefix(parentId, "ou-") {
		// Get OU name
		orgUnit, err := client.DescribeOrganizationalUnit(context.TODO(), &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: aws.String(parentId),
		})
		if err != nil {
			return nil, err
		}
		organization.Name = *orgUnit.OrganizationalUnit.Name
		organization.Type = "ORGANIZATIONAL_UNIT"
	} else if accountRegexp.MatchString(parentId) {
		// Get account name
		accountDetails, err := client.DescribeAccount(context.TODO(), &organizations.DescribeAccountInput{
			AccountId: aws.String(parentId),
		})
		if err != nil {
			return nil, err
		}
		organization.Name = *accountDetails.Account.Name
		organization.Type = "ACCOUNT"
		return organization, nil
	} else {
		return nil, errors.New("Unknown parent id format " + parentId)
	}

	// Create iterator to go over OU pages
	orgUnitPaginator := organizations.NewListChildrenPaginator(client, &organizations.ListChildrenInput{
		ParentId:  aws.String(parentId),
		ChildType: types.ChildTypeOrganizationalUnit,
	})
	for orgUnitPaginator.HasMorePages() {
		results, err := orgUnitPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		// Go through all child OUs on this page
		for _, childOrgUnit := range results.Children {
			children, err := GetChildren(*childOrgUnit.Id, client)
			if err != nil {
				return nil, err
			}
			organization.OrgUnits = append(organization.OrgUnits, *children)
		}
	}

	// Create iterator to go over account pages
	accountPaginator := organizations.NewListChildrenPaginator(client, &organizations.ListChildrenInput{
		ParentId:  aws.String(parentId),
		ChildType: types.ChildTypeAccount,
	})
	for accountPaginator.HasMorePages() {
		results, err := accountPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		// Go through all child accounts on this page
		for _, childAccount := range results.Children {
			children, err := GetChildren(*childAccount.Id, client)
			if err != nil {
				return nil, err
			}
			organization.Accounts = append(organization.Accounts, *children)
		}
	}

	return organization, nil
}

func contains(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}

	return false
}
