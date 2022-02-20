package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	scps "github.com/MichaelPalmer1/aws-scripts/go/org-scps/lib"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func main() {
	var mode, targetId, format string
	flag.StringVar(&mode, "mode", "all", "Fetch all policies or only policies effective for a particular target {all, effective}")
	flag.StringVar(&targetId, "target-id", "", "When --mode=effective, specify the target to fetch effective policies for")
	flag.StringVar(&format, "format", "json", "Specify the output format {json, file}")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg.Region = "us-east-1"
	orgs := organizations.NewFromConfig(cfg)

	if mode == "effective" && targetId == "" {
		fmt.Println("--target-id is required when --mode=effective")
		os.Exit(1)
	}

	switch mode {
	case "all":
		policies, err := scps.GetScps(orgs)
		if err != nil {
			panic(err)
		}

		switch format {
		case "json":
			bs, err := json.MarshalIndent(policies, "", "  ")
			if err != nil {
				panic(err)
			}

			fmt.Println(string(bs))
		case "file":
			if err := os.Mkdir("policies", 0755); err != nil {
				panic(err)
			}

			for name, policy := range policies {
				bs, err := json.MarshalIndent(policy.Content, "", "  ")
				if err != nil {
					panic(err)
				}

				if err := ioutil.WriteFile("policies/"+name+".json", bs, 0755); err != nil {
					panic(err)
				}
			}
		}
	case "effective":
		policyIds, err := scps.GetEffectiveScpIds(targetId, orgs)
		if err != nil {
			panic(err)
		}

		policies, err := scps.GetPolicies(policyIds, orgs)
		if err != nil {
			panic(err)
		}

		switch format {
		case "json":
			bs, err := json.MarshalIndent(policies, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(bs))
		case "file":
			if err := os.MkdirAll("policies/"+targetId, 0755); err != nil {
				panic(err)
			}

			for name, policy := range policies {
				bs, err := json.MarshalIndent(policy, "", "  ")
				if err != nil {
					panic(err)
				}
				if err := ioutil.WriteFile(fmt.Sprintf("policies/%s/%s.json", targetId, name), bs, 0755); err != nil {
					panic(err)
				}
			}
		}
	}

}
