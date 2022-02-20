package structure

import (
	"context"
	"errors"
	"strings"

	orgUtils "github.com/MichaelPalmer1/aws-scripts/go/org-utils"
	"github.com/MichaelPalmer1/aws-scripts/go/utils"
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
			policyName := aws.ToString(policy.Name)
			if !utils.Contains(policies, policyName) {
				policies = append(policies, policyName)
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
		organization.Name = aws.ToString(orgUnit.OrganizationalUnit.Name)
		organization.Type = "ORGANIZATIONAL_UNIT"
	} else if orgUtils.AccountRegex.MatchString(parentId) {
		// Get account name
		accountDetails, err := client.DescribeAccount(context.TODO(), &organizations.DescribeAccountInput{
			AccountId: aws.String(parentId),
		})
		if err != nil {
			return nil, err
		}
		organization.Name = aws.ToString(accountDetails.Account.Name)
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
			children, err := GetChildren(aws.ToString(childOrgUnit.Id), client)
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
			children, err := GetChildren(aws.ToString(childAccount.Id), client)
			if err != nil {
				return nil, err
			}
			organization.Accounts = append(organization.Accounts, *children)
		}
	}

	return organization, nil
}
