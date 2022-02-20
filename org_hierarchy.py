from argparse import ArgumentParser
from utils import build_list
import boto3
import json

orgs = boto3.client('organizations')

def get_hierarchy(child_id: str):
    hierarchy = []
    if child_id.startswith('r-'):
        return [{
            'Id': child_id,
            'Name': 'Root',
            'Type': 'ROOT'
        }]
    elif child_id.startswith('ou-'):
        # Get OU name
        org_unit = orgs.describe_organizational_unit(OrganizationalUnitId=child_id)['OrganizationalUnit']
        hierarchy.append({
            'Id': child_id,
            'Name': org_unit['Name'],
            'Type': 'ORGANIZATIONAL_UNIT'
        })
    elif child_id.isnumeric():
        # Get account name
        account_details = orgs.describe_account(AccountId=child_id)['Account']
        hierarchy.append({
            'Id': child_id,
            'Name': account_details['Name'],
            'Type': 'ACCOUNT'
        })
    else:
        raise Exception(f'Unknown child id format {child_id}')

    # Get parents of the child
    parents = build_list(orgs.list_parents, ChildId=child_id)['Parents']
    hierarchy += get_hierarchy(parents[0]['Id'])

    # Reverse the output
    return hierarchy[::-1]


if __name__ == '__main__':
    # Create parser
    parser = ArgumentParser()
    parser.add_argument('--show-ids', action='store_true', help='Whether to include OU IDs')
    parser.add_argument('--format', choices=('text', 'json'), default='text')
    parser.add_argument('account', type=str, help='AWS Account ID or Name')
    opts = parser.parse_args()

    # Get account id from organization
    account_id = None
    if opts.account.isnumeric():
        account_id = opts.account
    else:
        accounts = build_list(orgs.list_accounts)

        for account in accounts['Accounts']:
            if account['Name'] == opts.account:
                account_id = account['Id']
                break
        
        if not account_id:
            raise Exception(f'Account "{opts.account}" does not exist in this organization')

    # Get hierarchy
    hierarchy = get_hierarchy(account_id)

    # Output formats
    if opts.format == 'text':
        if opts.show_ids:
            print(' -> '.join(map(lambda x: f'{x["Name"]} ({x["Id"]})', hierarchy)))
        else:
            print(' -> '.join(map(lambda x: x['Name'], hierarchy)))
    
    elif opts.format == 'json':
        print(json.dumps(hierarchy, indent=4))
