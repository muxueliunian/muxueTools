#!/usr/bin/env python3
"""
PNG to ICO converter for MuxueTools
Converts gugugaga.png to a multi-size .ico file
Supports transparent background
"""

import sys
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("‚ù?Pillow library not found!")
    print("\nPlease install it with:")
    print("  pip install Pillow")
    sys.exit(1)

def create_ico(input_png: str, output_ico: str):
    """Convert PNG to ICO with multiple sizes"""
    
    print("MuxueTools Icon Converter")
    print("=" * 40)
    
    # Check input
    input_path = Path(input_png)
    if not input_path.exists():
        print(f"‚ù?Input file not found: {input_png}")
        sys.exit(1)
    
    print(f"üìÅ Input:  {input_path}")
    
    # Load PNG
    try:
        img = Image.open(input_path)
        print(f"‚ú?Loaded: {img.size[0]}x{img.size[1]} {img.mode}")
    except Exception as e:
        print(f"‚ù?Failed to load image: {e}")
        sys.exit(1)
    
    # Create output directory
    output_path = Path(output_ico)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    # Standard Windows icon sizes
    sizes = [(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (16, 16)]
    
    print(f"\nüîÑ Generating {len(sizes)} icon sizes...")
    
    # Generate resized versions for ICO
    icon_sizes = []
    for size in sizes:
        resized = img.resize(size, Image.Resampling.LANCZOS)
        # Convert to RGBA if not already (preserve transparency)
        if resized.mode != 'RGBA':
            resized = resized.convert('RGBA')
        icon_sizes.append(resized)
        print(f"  ‚ú?{size[0]}x{size[1]}")
    
    # Save as ICO
    try:
        icon_sizes[0].save(
            output_path,
            format='ICO',
            sizes=[s.size for s in icon_sizes],
            append_images=icon_sizes[1:]
        )
        print(f"\n‚ú?Icon created: {output_path}")
        
        file_size = output_path.stat().st_size
        print(f"  Size: {file_size / 1024:.2f} KB")
        
    except Exception as e:
        print(f"‚ù?Failed to save ICO: {e}")
        sys.exit(1)
    
    print("\n‚ú?Conversion complete!")

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Convert PNG to multi-size ICO for Windows")
    parser.add_argument("-i", "--input", default="image/gugugaga-removebg-preview.png",
                        help="Input PNG file path (default: image/gugugaga-removebg-preview.png)")
    parser.add_argument("-o", "--output", default="assets/icon.ico",
                        help="Output ICO file path (default: assets/icon.ico)")
    
    args = parser.parse_args()
    create_ico(args.input, args.output)
