import boto3
import json
from argparse import ArgumentParser
from diagrams import Diagram, Node
from diagrams.aws.management import Organizations, OrganizationsAccount, OrganizationsOrganizationalUnit
from utils import build_list

orgs = boto3.client('organizations')

def get_policies(parent_id: str):
    policies = orgs.list_policies_for_target(TargetId=parent_id, Filter='SERVICE_CONTROL_POLICY')
    return [
        policy['Name'] for policy in policies['Policies']
    ]

def get_children(parent_id: str):
    organization = {'Policies': get_policies(parent_id)}

    if parent_id.startswith('r-'):
        organization.update({
            'Id': parent_id,
            'Name': 'Root',
            'Type': 'ROOT'
        })

    elif parent_id.startswith('ou-'):
        # Get OU name
        org_unit = orgs.describe_organizational_unit(OrganizationalUnitId=parent_id)['OrganizationalUnit']
        organization.update({
            'Id': parent_id,
            'Name': org_unit['Name'],
            'Type': 'ORGANIZATIONAL_UNIT'
        })

    elif parent_id.isnumeric():
        # Get account name
        account_details = orgs.describe_account(AccountId=parent_id)['Account']
        organization.update({
            'Id': parent_id,
            'Name': account_details['Name'],
            'Type': 'ACCOUNT'
        })
        
        return organization
    else:
        raise Exception(f'Unknown parent id format {parent_id}')

    organization.update({
        'OrgUnits': [],
        'Accounts': []
    })

    child_ous = build_list(orgs.list_children, ParentId=parent_id, ChildType='ORGANIZATIONAL_UNIT')['Children']
    for child_ou in child_ous:
        organization['OrgUnits'].append(get_children(child_ou['Id']))

    child_accounts = build_list(orgs.list_children, ParentId=parent_id, ChildType='ACCOUNT')['Children']
    for child_account in child_accounts:
        organization['Accounts'].append(get_children(child_account['Id']))

    return organization

def build_diagram(org_structure: dict, show_ids=False, show_accounts=True, show_policies=False):
    with Diagram(filename='organization', direction='TB', curvestyle='curved', show=False):
        def render_child(child_structure: dict, parent_node: Node = None):
            # Detect node type
            if child_structure['Type'] == 'ROOT':
                node_type = Organizations
            elif child_structure['Type'] == 'ORGANIZATIONAL_UNIT':
                node_type = OrganizationsOrganizationalUnit
            elif child_structure['Type'] == 'ACCOUNT':
                node_type = OrganizationsAccount
                if not show_accounts:
                    return
            else:
                raise Exception(f'Unknown node type {child_structure["Type"]}')
            
            # Build node
            text = child_structure['Name']
            if show_ids:
                text += f'\n({child_structure["Id"]})'
            if show_policies:
                text += f'\n[{", ".join(child_structure["Policies"])}]'
            node = node_type(text)

            # Create relationship
            if parent_node:
                parent_node >> node

            # Render child org units
            for org_unit in child_structure.get('OrgUnits', []):
                render_child(org_unit, node)
            
            # Render child accounts
            if show_accounts:
                for account in child_structure.get('Accounts', []):
                    render_child(account, node)

        render_child(org_structure)

def print_structure(org_structure, depth=0, show_ids=False, show_accounts=True, show_policies=False):
        # Build text
        text = ''
        if org_structure['Type'] == 'ROOT':
            text += '\x1B[1m'
        elif org_structure['Type'] == 'ORGANIZATIONAL_UNIT':
            text += '\x1B[4m'
        elif org_structure['Type'] == 'ACCOUNT':
            text += '\x1B[3m'

        text += org_structure['Name']
        text += '\x1B[0m'

        if show_ids:
            text += f' ({org_structure["Id"]})'        
        if show_policies:
            text += f'\t[{", ".join(org_structure["Policies"])}]'
        
        if depth == 0 and org_structure['Type'] == 'ROOT':
            print(text)
        elif depth == 1:
            print('|-- ' + text)
        else:
            print('  ' * depth + '|-- ' + text)

        # Render child org units
        for org_unit in org_structure.get('OrgUnits', []):
            print_structure(org_unit, depth=depth + 1, show_ids=show_ids, show_accounts=show_accounts, show_policies=show_policies)
        
        # Render child accounts
        if show_accounts:
            for account in org_structure.get('Accounts', []):
                print_structure(account, depth=depth + 1, show_ids=show_ids, show_accounts=show_accounts, show_policies=show_policies)

if __name__ == '__main__':
    # Create parser
    parser = ArgumentParser()
    parser.add_argument('--show-ids', action='store_true', help='Show IDs in the output')
    parser.add_argument('--show-accounts', action='store_true', help='Show accounts in the output')
    parser.add_argument('--show-policies', action='store_true', help='Show SCPs in the output')
    parser.add_argument('--format', choices=('png', 'json', 'text'), default='text')
    opts = parser.parse_args()

    roots = orgs.list_roots()
    organization = get_children(roots['Roots'][0]['Id'])

    if opts.format == 'json':
        print(json.dumps(organization, indent=4))
    elif opts.format == 'text':
        print_structure(organization, show_ids=opts.show_ids, show_accounts=opts.show_accounts, show_policies=opts.show_policies) 
        print('\nLegend: \x1B[1mROOT\x1B[0m\t\x1B[4mOU\x1B[0m\t\x1B[3mACCOUNT\x1B[0m')
    elif opts.format == 'png':
        build_diagram(organization, show_ids=opts.show_ids, show_accounts=opts.show_accounts, show_policies=opts.show_policies)
