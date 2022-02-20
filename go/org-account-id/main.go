package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

var orgs *organizations.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg.Region = "us-east-1"
	orgs = organizations.NewFromConfig(cfg)
}

func GetAccountId(accountName string) *string {
	paginator := organizations.NewListAccountsPaginator(orgs, &organizations.ListAccountsInput{})
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

func main() {
	accountName := flag.String("account-name", "", "AWS Account Name")
	flag.Parse()

	accountId := GetAccountId(*accountName)
	if accountId == nil {
		fmt.Printf("Account name \"%s\" does not exist in this organization\n", *accountName)
		os.Exit(1)
	}

	fmt.Println(*accountId)
}
