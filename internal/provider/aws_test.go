package provider

import (
	"testing"
)

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		prefix string
		want   string
	}{
		{
			name:   "strips matching prefix",
			s:      "/myapp/prod/DB_HOST",
			prefix: "/myapp/prod/",
			want:   "DB_HOST",
		},
		{
			name:   "no match returns original",
			s:      "/other/DB_HOST",
			prefix: "/myapp/prod/",
			want:   "/other/DB_HOST",
		},
		{
			name:   "exact prefix no suffix returns original",
			s:      "/myapp/prod/",
			prefix: "/myapp/prod/",
			want:   "/myapp/prod/",
		},
		{
			name:   "empty prefix returns original",
			s:      "DB_HOST",
			prefix: "",
			want:   "DB_HOST",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stripPrefix(tc.s, tc.prefix)
			if got != tc.want {
				t.Errorf("stripPrefix(%q, %q) = %q; want %q", tc.s, tc.prefix, got, tc.want)
			}
		})
	}
}

func TestAWSProvider_Name(t *testing.T) {
	p := &AWSProvider{pathPrefix: "/app/"}
	if got := p.Name(); got != "aws-ssm" {
		t.Errorf("Name() = %q; want %q", got, "aws-ssm")
	}
}
