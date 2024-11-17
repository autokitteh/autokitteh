import json


def print_json(label, data):
    print(label + ":\n" + json.dumps(data, indent=2))
