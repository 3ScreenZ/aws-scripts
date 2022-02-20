from argparse import ArgumentParser
import boto3

orgs = boto3.client('organizations')
account_paginator = orgs.get_paginator('list_accounts')

def get_account_id(account_name):
    for page in account_paginator.paginate():
        for account in page['Accounts']:
            if account['Name'] == account_name:
                return account['Id']
    
    raise Exception(f'Account with name "{account_name}" does not exist in this organization')

if __name__ == '__main__':
    parser = ArgumentParser()
    parser.add_argument('account_name', type=str, help='Account name')
    opts = parser.parse_args()

    # Get account id
    account_id = get_account_id(opts.account_name)
    print(account_id)
