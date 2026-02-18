import os
import csv

DOCS_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "ocrflow", "store", "items_metadata")

ENTRIES_CSVS = [
    os.path.join(DOCS_DIR, 'items_print.csv')
]

CITIES_CSV = os.path.join(DOCS_DIR, 'cities.csv')

cities_from_csvs = set()

for csv_file in ENTRIES_CSVS:
    with open(csv_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            if row['city']:
                for city in row['city'].split(','):
                    cities_from_csvs.add(city.strip())

cities_in_reference = set()

with open(CITIES_CSV, 'r', encoding='utf-8') as f:
    reader = csv.DictReader(f)
    for row in reader:
        cities_in_reference.add(row['city'])

missing_cities = cities_from_csvs - cities_in_reference

for city in sorted(missing_cities):
    print(city)
