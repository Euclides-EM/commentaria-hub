# Euclid's Elements: A Resource Box

Created and maintained by Mia Joskowicz (miajosko@gmail.com)

This digital humanities project explores the early modern transmission of Euclid's _Elements_ through an interactive web application that serves as a living resource book accompanying PhD research in this area. The project provides multiple pathways to explore editions of Euclidean and related mathematical texts from 1482-1703.

## About the Project

This web application combines traditional scholarly research with digital tools to analyze the circulation and transmission of Euclid's _Elements_ during the early modern period. The project is rooted in the understanding that title pages serve multiple roles: they designate a book's identity, function as instruments of intellectual and commercial positioning, reflect contemporary aesthetic and typographical conventions, and encode broader intellectual and disciplinary priorities.

## Web App Features

### Core Navigation

- **Catalog**: Browse editions in a structured table format with key bibliographic details and links to digital facsimiles when available
- **Gallery**: View the same collection with thumbnails and expandable detailed views of each edition
- **Map**: Explore editions geographically placed along an interactive timeline
- **Trends**: Analyze publication patterns and trends across time and geography

### Title Pages Experiment

The application includes an experimental analysis of title pages from the corpus. This feature segments title page text into distinct elements to identify patterns and variations across different editions, publishers, and time periods. The transcription and segmentation were performed using a combination of automated processes and large language models.

### Filtering System

All sections share a unified filtering system that preserves search parameters as users navigate between different views, enabling consistent exploration across the dataset.

## Data Sources and Credits

The corpus is based primarily on the [Euclid in print (1482–1703) catalog](https://bibsoc.org.uk/euclid-print-1482-1703/), supplemented with references from scholarship, USTC, the BnF collection, and Google Books.

### Dataset Structure

The project's data is organized in the `public/` directory:

- **`docs/*.csv`**: Core dataset files containing bibliographic and analytical data
  - `EiP.csv`: Main catalog of editions with comprehensive metadata including bibliographic details, title page transcriptions, and analytical features
  - `cities.csv`: Geographic data with historical political boundaries for mapping functionality
  - `dotted_1.csv` and `dotted_2.csv`: Supplementary datasets for specific analyses
  - `EiP-secondary.csv`: Additional reference materials
- **`tps/`**: Title page images organized by city and year (e.g., `Paris_1622_tp.jpeg`)
- **`presentation/`**: Academic presentation materials
- **Static assets**: Fonts, icons, and interface images

### Data Methodology

The CSV datasets combine manual scholarly curation with automated enrichment:

- **Manual curation**: Core bibliographic properties, classifications, and scholarly annotations were manually researched and entered
- **LLM enrichment**: Title page transcriptions and certain analytical features were generated using the scripts in the `enricher/` directory, which employ large language models for text analysis and feature extraction
- **Hybrid approach**: The combination ensures scholarly accuracy for foundational data while leveraging computational methods for detailed textual analysis

## Deployment

The application is deployed and accessible at: https://elements-resource-box.netlify.app/

## Technical Requirements

### Enricher Component

The project includes Python scripts for data enrichment and analysis located in the `enricher/` directory. To run the enricher code, you must provide an OpenAI API token via the `OPENAI_API_KEY` environment variable.

### Development

#### Running the Web Application

```bash
# Install dependencies
yarn install

# Start development server
yarn dev

# Build for production
yarn build

# Preview production build
yarn preview
```

#### Running the Enricher Scripts

The enricher scripts require Python 3.12+ and use `uv` for dependency management:

```bash
# Navigate to enricher directory
cd enricher

# Install dependencies with uv
uv sync

# Set API keys
export OPENAI_API_KEY=your_api_key_here
export GOOGLE_APPLICATION_CREDENTIALS=path/to/your/google-credentials.json

# Run enricher scripts
uv run python enricher.py
```

Additional enricher scripts include:

- `transcribe.py`: Title page transcription
- `translator.py`: Text translation services
- `verifier.py`: Data validation and verification

#### Technology Stack

- Built with React and TypeScript
- Uses Vite for development and build processes
- Styled with Emotion for component styling
- Python enricher scripts use OpenAI API and optionally Google Cloud Translate

## License

This project is under the [CC BY-NC (4.0) license/https://creativecommons.org/licenses/by-nc/4.0/]. Permission is granted to copy, transmit and reuse the text and its contents, subject only to the conditions that the original authors are acknowledged and that no commercial use is made of it.
