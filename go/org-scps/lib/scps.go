package scps

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type SCP struct {
	Summary types.PolicySummary
	Content interface{}
}

func GetEffectiveScpIds(targetId string, client *organizations.Client) ([]string, error) {
	var policyIds []string

	paginator := organizations.NewListPoliciesForTargetPaginator(client, &organizations.ListPoliciesForTargetInput{
		TargetId: aws.String(targetId),
		Filter:   types.PolicyTypeServiceControlPolicy,
	})

	for paginator.HasMorePages() {
		results, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, policy := range results.Policies {
			if !contains(policyIds, *policy.Id) {
				policyIds = append(policyIds, *policy.Id)
			}
		}
	}

	// Skip root
	if !strings.HasPrefix(targetId, "r-") {
		parents, err := client.ListParents(context.TODO(), &organizations.ListParentsInput{
			ChildId: aws.String(targetId),
		})
		if err != nil {
			return nil, err
		}

		// Get SCPs on parent
		parentId := *parents.Parents[0].Id
		parentScpIds, err := GetEffectiveScpIds(parentId, client)
		if err != nil {
			return nil, err
		}
		policyIds = append(policyIds, parentScpIds...)
	}

	return policyIds, nil
}

func GetPolicies(policyIds []string, client *organizations.Client) (map[string]interface{}, error) {
	policies := make(map[string]interface{})

	for _, policyId := range policyIds {
		policy, err := client.DescribePolicy(context.TODO(), &organizations.DescribePolicyInput{
			PolicyId: aws.String(policyId),
		})
		if err != nil {
			return nil, err
		}

		var policyContent interface{}
		if err := json.Unmarshal([]byte(*policy.Policy.Content), &policyContent); err != nil {
			return nil, err
		}

		policies[*policy.Policy.PolicySummary.Name] = policyContent
	}

	return policies, nil
}

func GetScps(client *organizations.Client) (map[string]SCP, error) {
	policies := make(map[string]SCP)
	paginator := organizations.NewListPoliciesPaginator(client, &organizations.ListPoliciesInput{
		Filter: types.PolicyTypeServiceControlPolicy,
	})

	for paginator.HasMorePages() {
		results, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, result := range results.Policies {
			var policyContent interface{}
			policyDetails, err := client.DescribePolicy(context.TODO(), &organizations.DescribePolicyInput{
				PolicyId: result.Id,
			})
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal([]byte(*policyDetails.Policy.Content), &policyContent); err != nil {
				return nil, err
			}

			policies[*result.Name] = SCP{
				Summary: result,
				Content: policyContent,
			}
		}
	}

	return policies, nil
}

func contains(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}

	return false
}
