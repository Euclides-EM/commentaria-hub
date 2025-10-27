from pdf2image import convert_from_path
import os
import sys

def pdf_to_jpg(pdf_path, output_dir=None, dpi=300, start_page=None, end_page=None):
    if not os.path.exists(pdf_path):
        raise FileNotFoundError(f"PDF not found: {pdf_path}")

    if output_dir is None:
        output_dir = os.path.splitext(pdf_path)[0] + "_pages"
    os.makedirs(output_dir, exist_ok=True)

    print(f"Converting '{pdf_path}' to JPEG images...")
    images = convert_from_path(
        pdf_path,
        dpi=dpi,
        first_page=start_page,
        last_page=end_page
    )

    # Adjust numbering offset if starting from a specific page
    offset = start_page - 1 if start_page else 0

    for i, image in enumerate(images, start=1):
        page_num = i + offset
        jpg_path = os.path.join(output_dir, f"page_{page_num:03}.jpg")
        image.save(jpg_path, "JPEG")
        print(f"Saved {jpg_path}")

    print(f"Done. Images saved in '{output_dir}'")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python pdf_to_jpg.py input.pdf [output_dir] [start_page] [end_page]")
        print("Example: python pdf_to_jpg.py myfile.pdf output 3 7")
        sys.exit(1)

    pdf_file = sys.argv[1]
    output_dir = sys.argv[2] if len(sys.argv) > 2 and sys.argv[2].isdigit() is False else None
    start_page = int(sys.argv[3 if output_dir else 2]) if len(sys.argv) > (3 if output_dir else 2) else None
    end_page = int(sys.argv[4 if output_dir else 3]) if len(sys.argv) > (4 if output_dir else 3) else None

    pdf_to_jpg(pdf_file, output_dir, start_page=start_page, end_page=end_page)
