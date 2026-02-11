package okeycrawler

import "testing"

func Test_okeyCrawler_checkPageType(t *testing.T) {
	crawler := NewOkeyCrawler()

	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "major category page",
			html: `<div class="rows categories"></div>`,
			want: majorPageType,
		},
		{
			name: "minor category page",
			html: `<ul class="grid_mode grid rows"></ul>`,
			want: minorPageType,
		},
		{
			name: "main product page",
			html: `<div class="rows product_main_info"></div>`,
			want: mainProductPageType,
		},
		{
			name: "unknown page",
			html: `<html><body>something else</body></html>`,
			want: unknownPageType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := crawler.checkPageType(tt.html)
			if got != tt.want {
				t.Fatalf("checkPageType() = %q, want %q", got, tt.want)
			}
		})
	}
}
