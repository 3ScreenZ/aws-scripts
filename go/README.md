# AWS Scripts - Go

## [Organization Hierarchy](org-hierarchy/main.go)

Detects hierarchy of an under within an AWS Organization.

```
Root (r-abc1) -> product-a (ou-abc1-abcdef12) -> sku-a (ou-abc1-abcdef22) -> dev (ou-abc1-abcdef33) -> account-1 (111111111111)
```

```json
[
    {
        "Id": "r-abc1",
        "Name": "Root",
        "Type": "ROOT"
    },
    {
        "Id": "ou-abc1-abcdef12",
        "Name": "product-a",
        "Type": "ORGANIZATIONAL_UNIT"
    },
    {
        "Id": "ou-abc1-abcdef22",
        "Name": "sku-a",
        "Type": "ORGANIZATIONAL_UNIT"
    },
    {
        "Id": "ou-abc1-abcdef33",
        "Name": "dev",
        "Type": "ORGANIZATIONAL_UNIT"
    },
    {
        "Id": "111111111111",
        "Name": "account-1",
        "Type": "ACCOUNT"
    }
]
```

## [Organization Structure](org-structure/main.go)

Renders structure of an AWS Organization. Supports text, json, and png outputs.

```
Root (r-abc1)
|-- product-a (ou-abc1-abcdef12)
    |-- sku-a (ou-abc1-abcdef22)
        |-- dev (ou-abc1-abcdef33)
            |-- account-1 (111111111111)
            |-- account-2 (222222222222)
            |-- account-3 (333333333333)
        |-- stage (ou-abc1-abcdef44)
            |-- account-4 (444444444444)
        |-- prod  (ou-abc1-abcdef55)
            |-- account-5 (555555555555)
            |-- account-6 (666666666666)
    |-- sku-b (ou-abc1-abc12fde)
        |-- dev (ou-abc1-abc42fde)
        |-- stage (ou-abc1-abc12fae)
            |-- account-7 (777777777777)
        |-- prod (ou-abc1-acc12fde)
            |-- account-8 (888888888888)
|-- product-b (ou-abc1-bbc12fde)
    |-- sku-a (ou-abc1-ccc12fde)
    |-- sku-b (ou-abc1-aaa12fde)
|-- account-9 (999999999999)
```

```json
{
    "Id": "r-abc1",
    "Name": "Root",
    "Type": "ROOT",
    "OrgUnits": [
        {
            "Id": "ou-abc1-abcdef12",
            "Name": "product-a",
            "Type": "ORGANIZATIONAL_UNIT",
            "OrgUnits": [
                {
                    "Id": "ou-abc1-abcdef22",
                    "Name": "sku-a",
                    "Type": "ORGANIZATIONAL_UNIT",
                    "OrgUnits": [
                        {
                            "Id": "ou-abc1-abcdef33",
                            "Name": "dev",
                            "Type": "ORGANIZATIONAL_UNIT",
                            "OrgUnits": [],
                            "Accounts": [
                                {
                                    "Id": "111111111111",
                                    "Name": "account-1",
                                    "Type": "ACCOUNT"
                                },
                                {
                                    "Id": "222222222222",
                                    "Name": "account-2",
                                    "Type": "ACCOUNT"
                                },
                                {
                                    "Id": "333333333333",
                                    "Name": "account-3",
                                    "Type": "ACCOUNT"
                                }
                            ]
                        }
                    ],
                    "Accounts": []
                }
            ],
            "Accounts": []
        }
    ],
    "Accounts": [
        {
            "Id": "999999999999",
            "Name": "account-9",
            "Type": "ACCOUNT"
        }
    ]
}
```

## [Organization SCPs](org-scps/main.go)

Fetch SCPs for an Organization

## [Organization Account Id](org-account-id/main.go)

Given an account name, get its account id from Organizations
