from argparse import ArgumentParser
from utils import build_list
import boto3
import json
import os

orgs = boto3.client('organizations')
policy_paginator = orgs.get_paginator('list_policies')
target_paginator = orgs.get_paginator('list_policies_for_target')

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

def get_effective_scp_ids(target_id):
    policy_ids = set()

    paginator = target_paginator.paginate(TargetId=target_id, Filter='SERVICE_CONTROL_POLICY')
    for page in paginator:
        for policy in page['Policies']:
            policy_ids.add(policy['Id'])
    
    # Skip root
    if not target_id.startswith('r-'):
        parents = build_list(orgs.list_parents, ChildId=target_id)
        parent_id = parents['Parents'][0]['Id']
        policy_ids |= get_effective_scp_ids(parent_id)

    return policy_ids

def get_policies(policy_ids):
    policies = {}

    for policy_id in policy_ids:
        policy = orgs.describe_policy(PolicyId=policy_id)['Policy']
        policies[policy['PolicySummary']['Name']] = json.loads(policy['Content'])
    
    return policies

if __name__ == '__main__':
    parser = ArgumentParser(description='Fetch information about AWS Organizations service control policies')
    parser.add_argument('--mode', type=str, choices=('all', 'effective'), default='all', help='Fetch all policies or only policies effective for a particular target')
    parser.add_argument('--target-id', type=str, help='When --mode=effective, specify the target to fetch effective policies for')
    parser.add_argument('--format', type=str, choices=('json', 'file'), default='json', help='Specify the output format')
    opts = parser.parse_args()
    if opts.mode == 'effective' and not opts.target_id:
        parser.error('--target-id is required when --mode=effective')

    if opts.mode == 'all':
        policies = get_scps()
        if opts.format == 'json':
            print(json.dumps(policies, indent=4))
        elif opts.format == 'file':
            os.makedirs('policies', exist_ok=True)
            for name, policy in policies.items():
                with open(f'policies/{name}.json', 'w') as f:
                    json.dump(policy['Content'], f, indent=4)
    elif opts.mode == 'effective':
        policy_ids = get_effective_scp_ids(opts.target_id)
        policies = get_policies(policy_ids)

        if opts.format == 'json':
            print(json.dumps(policies, indent=4))
        elif opts.format == 'file':
            os.makedirs(f'policies/{opts.target_id}', exist_ok=True)
            for name, policy in policies.items():
                with open(f'policies/{opts.target_id}/{name}.json', 'w') as f:
                    json.dump(policy, f, indent=4)
