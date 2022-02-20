package account

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func GetAccountId(accountName string, client *organizations.Client) *string {
	paginator := organizations.NewListAccountsPaginator(client, &organizations.ListAccountsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			panic(err)
		}

		for _, account := range output.Accounts {
			if *account.Name == accountName {
				return account.Id
			}
		}
	}

	return nil
}
