#include <cairo.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

double width;
double height;
double margin;
double line_distance;
double white_width;
double blue_width;

void red(cairo_t *cr)
{
	cairo_set_source_rgba(cr,
		0xca / 256.0,
		0x2b / 256.0,
		0x2a / 256.0,
		1.0);
}

void white(cairo_t *cr)
{
	cairo_set_source_rgba(cr, 1.0, 1.0, 1.0, 1.0);
	cairo_set_line_width(cr, white_width);
	cairo_set_line_cap(cr, CAIRO_LINE_CAP_SQUARE);
}

void blue(cairo_t *cr)
{

	cairo_set_source_rgba(cr,
		0x66 / 256.0,
		0x55 / 256.0,
		0xbb / 256.0,
		1.0);
	cairo_set_line_width(cr, blue_width);
	cairo_set_line_cap(cr, CAIRO_LINE_CAP_SQUARE);
}

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

	margin = w / 500.0 * 100.0;
	line_distance = w / 500.0 * 150.0;
	white_width = w / 500.0 * 80.0;
	blue_width = w / 500.0 * 60.0;

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
	red(cr);
	cairo_paint(cr);

	static cairo_matrix_t m = {
		.xx = 1.0,
		.yx = 0.0,
		.xy = -0.2,
		.yy = 1.0,
		.x0 = 50.0,
		.y0 = 0.0,
	};
	m.x0 = w / 500.0 * 50.0;
	cairo_set_matrix(cr, &m);

	/* Drawing horizontal white lines */
	white(cr);
	cairo_move_to(cr, margin, height / 2.0 - line_distance / 2.0);
	cairo_rel_line_to(cr, width - 2 * margin, 0);
	cairo_move_to(cr, margin, height / 2.0 + line_distance / 2.0);
	cairo_rel_line_to(cr, width - 2 * margin, 0);
	cairo_stroke(cr);

	/* Drawing vertical white lines */
	white(cr);
	cairo_move_to(cr, width / 2.0 - line_distance / 2.0, margin);
	cairo_rel_line_to(cr, 0, height - margin * 2);
	cairo_move_to(cr, width / 2.0 + line_distance / 2.0, margin);
	cairo_rel_line_to(cr, 0, height - margin * 2);
	cairo_stroke(cr);

	/* Drawing horizontal blue lines */
	blue(cr);
	cairo_move_to(cr, margin, height / 2.0 - line_distance / 2.0);
	cairo_rel_line_to(cr, width - 2 * margin, 0);
	cairo_move_to(cr, margin, height / 2.0 + line_distance / 2.0);
	cairo_rel_line_to(cr, width - 2 * margin, 0);
	cairo_stroke(cr);
	
	/* Drawing vertical blue lines */
	blue(cr);
	cairo_move_to(cr, width / 2.0 - line_distance / 2.0, margin);
	cairo_rel_line_to(cr, 0, height - margin * 2);
	cairo_move_to(cr, width / 2.0 + line_distance / 2.0, margin);
	cairo_rel_line_to(cr, 0, height - margin * 2);
	cairo_stroke(cr);

	cairo_destroy(cr);

	if(imgtype == TYPE_PNG)
		cairo_surface_write_to_png(surface, output_filename);

	cairo_surface_destroy(surface);
	exit(0);
}
