package hierarchy

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

type Child struct {
	Id   string
	Name string
	Type string
}

var AccountRegex *regexp.Regexp

func init() {
	AccountRegex = regexp.MustCompile(`\d{12}`)
}

func GetHierarchy(childId string, client *organizations.Client) ([]Child, error) {
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
		ouOutput, err := client.DescribeOrganizationalUnit(context.TODO(), &organizations.DescribeOrganizationalUnitInput{
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
	} else if AccountRegex.MatchString(childId) {
		acctOutput, err := client.DescribeAccount(context.TODO(), &organizations.DescribeAccountInput{
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

	parentOutput, err := client.ListParents(context.TODO(), &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	})
	if err != nil {
		return nil, err
	}

	childHierarchy, err := GetHierarchy(*parentOutput.Parents[0].Id, client)
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
