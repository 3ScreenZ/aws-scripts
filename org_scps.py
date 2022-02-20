from argparse import ArgumentParser
import boto3
import json
import os

orgs = boto3.client('organizations')
policy_paginator = orgs.get_paginator('list_policies')

def get_scps():
    policies = {}
    paginator = policy_paginator.paginate(Filter='SERVICE_CONTROL_POLICY')
    
    for page in paginator:
        for item in page['Policies']:
            content = orgs.describe_policy(PolicyId=item['Id'])['Policy']

            policies[content['PolicySummary']['Name']] = {
                'Summary': content['PolicySummary'],
                'Content': json.loads(content['Content'])
            }
    
    return policies


if __name__ == '__main__':
    parser = ArgumentParser()
    parser.add_argument('--format', type=str, choices=('json', 'files'), default='json')
    opts = parser.parse_args()

    policies = get_scps()

    if opts.format == 'json':
        print(json.dumps(policies, indent=4))
    elif opts.format == 'files':
        os.makedirs('policies', exist_ok=True)
        for name, policy in policies.items():
            with open(f'policies/{name}.json', 'w') as f:
                json.dump(policy['Content'], f, indent=4)
