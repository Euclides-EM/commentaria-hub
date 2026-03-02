import os

from kraken import rpred
from kraken.lib import models, xml

MODEL_PATH = os.environ["MODEL_PATH"]
INPUT_ALTO_PATH = os.environ["INPUT_ALTO_PATH"]
IMAGE_PATH = os.environ["IMAGE_PATH"]

# 1. Load the model
model = models.load_any(MODEL_PATH)

# 2. Open your ALTO and parse the segments
with open(INPUT_ALTO_PATH, 'rb') as fp:
    # This specifically extracts the lines you already identified
    bounds = xml.parse_alto(fp)

# 3. Run OCR on those specific bounds
# 'bounds' tells Kraken exactly where the lines are on page-0131.png
it = rpred.rpred(model, IMAGE_PATH, bounds)

# 4. Save/Print results
for record in it:
    print(f"Line {record.index}: {record.prediction}")