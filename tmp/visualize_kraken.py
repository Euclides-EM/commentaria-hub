#!/usr/bin/env python3
"""
Visualize kraken baseline annotations on an image.
"""

import json
import sys
import matplotlib.pyplot as plt
import matplotlib.patches as patches
from PIL import Image
import numpy as np
import xml.etree.ElementTree as ET

def parse_points(points_str):
    """Parses a string of points "x1,y1 x2,y2..." into a list of [x, y] lists."""
    if not points_str:
        return []
    coords = []
    for pair in points_str.split():
        x, y = map(int, pair.split(','))
        coords.append([x, y])
    return coords

def load_kraken_data(data_file):
    """Load kraken annotation data (JSON or XML)."""
    try:
        with open(data_file, 'r') as f:
            data = json.load(f)
        return data
    except json.JSONDecodeError:
        # Not a JSON file, try parsing as XML
        try:
            tree = ET.parse(data_file)
            root = tree.getroot()
            lines_data = {'lines': []}
            # Assuming TextLine elements are nested within Page or TextRegion,
            # using .// to search all descendants
            for text_line in root.findall('.//TextLine'):
                baseline_elem = text_line.find('Baseline')
                coords_elem = text_line.find('Coords')

                baseline_points = []
                if baseline_elem is not None:
                    baseline_points = parse_points(baseline_elem.get('points'))

                boundary_points = []
                if coords_elem is not None:                   
                    boundary_points = parse_points(coords_elem.get('points'))

                lines_data['lines'].append({
                    'baseline': baseline_points,
                    'boundary': boundary_points
                })
            return lines_data
        except ET.ParseError as e:
            raise ValueError(f"Could not parse data file \'{data_file}\' as JSON or XML: {e}")

def visualize_baselines(image_file, data_file, output_file=None):
    """
    Visualize baselines and boundaries on the image.
    
    Args:
        image_file: Path to the image file
        data_file: Path to the kraken data file (JSON or XML)
        output_file: Optional path to save the visualization
    """
    # Load image
    img = Image.open(image_file)
    
    # Load kraken data
    data = load_kraken_data(data_file)
    
    # Create figure and axis
    fig, ax = plt.subplots(1, 1, figsize=(15, 20))
    ax.imshow(img)
    
    # Plot each line
    for i, line in enumerate(data['lines']):
        # Extract baseline coordinates
        baseline = line['baseline']
        boundary = line['boundary']
        
        # Plot baseline as a red line
        if baseline:
            baseline_x = [point[0] for point in baseline]
            baseline_y = [point[1] for point in baseline]
            ax.plot(baseline_x, baseline_y, 'r-', linewidth=2, alpha=0.8, label='Baseline' if i == 0 else "")
        
        # Plot boundary as a blue polygon
        if boundary:
            boundary_points = np.array(boundary)
            polygon = patches.Polygon(boundary_points, linewidth=1, edgecolor='blue', 
                                    facecolor='blue', alpha=0.2, label='Text Region' if i == 0 else "")
            ax.add_patch(polygon)
    
    # Set title and remove axes
    ax.set_title(f'Kraken Text Line Detection: {image_file}', fontsize=16)
    ax.axis('off')
    
    # Add legend
    ax.legend(loc='upper right')
    
    # Save or show
    if output_file:
        plt.savefig(output_file, dpi=150, bbox_inches='tight')
        print(f"Visualization saved to: {output_file}")
    else:
        plt.show()
    
    return fig, ax

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python visualize_kraken.py <image_file> <data_file> [output_file]")
        sys.exit(1)
    
    image_file = sys.argv[1]
    data_file = sys.argv[2]
    output_file = sys.argv[3] if len(sys.argv) > 3 else None
    
    try:
        visualize_baselines(image_file, data_file, output_file)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)