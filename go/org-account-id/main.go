package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	account "github.com/MichaelPalmer1/aws-scripts/go/org-account-id/lib"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func main() {
	accountName := flag.String("account-name", "", "AWS Account Name")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg.Region = "us-east-1"
	orgs := organizations.NewFromConfig(cfg)

	accountId := account.GetAccountId(*accountName, orgs)
	if accountId == nil {
		fmt.Printf("Account name \"%s\" does not exist in this organization\n", *accountName)
		os.Exit(1)
	}

	fmt.Println(*accountId)
}
