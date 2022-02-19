import random
from time import sleep
from botocore.exceptions import ClientError


def exponential_backoff(func, kwargs, error_codes=()):
    backoff_count = 0
    max_backoff_times = 8
    base_delay = 0.1
    delay = base_delay + random.uniform(0, 1)

    while True:
        try:
            return func(**kwargs)
        except ClientError as e:
            if e.response['Error']['Code'] == 'Throttling':
                max_backoff_times = 10
            elif e.response['Error']['Code'] not in error_codes:
                raise

            if backoff_count < max_backoff_times:
                sleep(delay)
                backoff_count += 1
                delay = base_delay * pow(2, backoff_count) + random.uniform(0, 1)
            else:
                raise

def build_list(func, **kwargs):
    if not isinstance(kwargs, dict):
        kwargs = {}

    items = {}
    response = exponential_backoff(func, kwargs, [])
    item_keys = []
    for key, value in response.items():
        if isinstance(value, list):
            item_keys.append(key)
            items[key] = response[key]
    while response.get('NextToken'):
        kwargs.update({'NextToken': response['NextToken']})
        response = exponential_backoff(func, kwargs, [])
        for key in item_keys:
            items[key].extend(response[key])
    return items
