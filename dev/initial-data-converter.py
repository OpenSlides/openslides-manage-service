"""
Download current initial-data.json from GitHub and run this script:

    $ cd dev
    $ wget https://raw.githubusercontent.com/OpenSlides/OpenSlides/master/docker/initial-data.json
    $ python initial-data-converter.py
    $ rm initial-data.json
    $ cd ..
"""

import json

INPUT_FILE = "initial-data.json"
OUTPUT_FILE = "../pkg/initialdata/default-initial-data.json"


def main():
    with open(INPUT_FILE) as initial_data_file:
        initial_data = json.load(initial_data_file)

    result = {}
    for collection in initial_data.keys():
        result[collection] = parse_collection(collection, initial_data[collection])

    with open(OUTPUT_FILE, "w") as output_file:
        json.dump(result, output_file, indent=4)
        output_file.write("\n")


def parse_collection(collection, value):
    if collection.startswith("_"):
        return value
    result = {}
    for obj in value:
        if collection == "user":
            obj["password"] = ""
        result[obj["id"]] = obj

    return result


if __name__ == "__main__":
    main()
    print("Done. Type git diff to see the changes.")
