#include <cairo.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

int width;
int height;

int main(int argc, char *argv[])
{
	enum {
		TYPE_PNG,
		TYPE_SVG,
	} imgtype = TYPE_SVG;

	if(argc != 4) {
		fprintf(stderr, "usage: ./gen <image-type> <width>x<height> <output-filename>\n");
		exit(1);
	}

	const char *output_filename;

	if(strcmp(argv[1], "png") == 0)
		imgtype = TYPE_PNG;
	if(strcmp(argv[1], "svg") == 0)
		imgtype = TYPE_SVG;

	unsigned w, h;
	int n = sscanf(argv[2], "%dx%d", &w, &h);
	if(n != 2) {
		fprintf(stderr, "Dimensions must be in the format of eg. 500x500.\n");
		exit(1);
	}
	width = w;
	height = h;

	output_filename = argv[3];

	cairo_surface_t *surface;
	switch(imgtype) {
	case TYPE_SVG:
		surface = cairo_svg_surface_create(output_filename, width, height);
		break;
	case TYPE_PNG:
		surface = cairo_image_surface_create(CAIRO_FORMAT_ARGB32, width, height);
		break;
	}

	cairo_status_t status = cairo_surface_status(surface);
	if(status != CAIRO_STATUS_SUCCESS) {
		fprintf(stderr, "Error on creating Cairo surface: %d.\n", status);
		exit(1);
	}

	cairo_t *cr = cairo_create(surface);

	cairo_set_source_rgba(cr, 0xca / 256.0, 0x2b / 256.0, 0x2a / 256.0, 1.0);
	cairo_paint(cr);

	cairo_select_font_face(cr, "Sujeta", CAIRO_FONT_SLANT_NORMAL, CAIRO_FONT_WEIGHT_NORMAL);
	cairo_set_font_size(cr, 60);
	cairo_move_to(cr, 1, height - 1);
	cairo_text_path(cr, "#Norske IRC-Kanaler");
	cairo_set_source_rgba(cr, 0x66 / 256.0, 0x55 / 256.0, 0xbb / 256.0, 1.0);
	cairo_fill_preserve(cr);
	cairo_set_source_rgba(cr, 1.0, 1.0, 1.0, 1.0);
	cairo_set_line_width(cr, 1.0);
	cairo_stroke(cr);

	cairo_destroy(cr);

	if(imgtype == TYPE_PNG)
		cairo_surface_write_to_png(surface, output_filename);

	cairo_surface_destroy(surface);
	exit(0);
}
