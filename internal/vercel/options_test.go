package vercel

import "testing"

func TestParseOptionsDefaultsToAccountMode(t *testing.T) {
	opts, err := ParseOptions(nil)
	if err != nil {
		t.Fatalf("ParseOptions returned error: %v", err)
	}

	if len(opts.Dirs) != 0 {
		t.Fatalf("expected no default pinned dirs, got %#v", opts.Dirs)
	}
	if opts.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", opts.Limit)
	}
	if opts.Refresh.String() != "30s" {
		t.Fatalf("expected default refresh 30s, got %s", opts.Refresh)
	}
}

func TestParseOptionsAcceptsRepeatedDirs(t *testing.T) {
	opts, err := ParseOptions([]string{"--dir", "/tmp/app", "--dir", "/tmp/admin", "--limit", "200", "--plain"})
	if err != nil {
		t.Fatalf("ParseOptions returned error: %v", err)
	}

	if len(opts.Dirs) != 2 {
		t.Fatalf("expected two dirs, got %#v", opts.Dirs)
	}
	if opts.Limit != 100 {
		t.Fatalf("expected limit to clamp to 100, got %d", opts.Limit)
	}
	if !opts.Plain {
		t.Fatal("expected plain mode")
	}
}

func TestParseOptionsRejectsInvalidLimit(t *testing.T) {
	_, err := ParseOptions([]string{"--limit", "0"})
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestParseOptionsRejectsNegativeRefresh(t *testing.T) {
	_, err := ParseOptions([]string{"--refresh", "-1s"})
	if err == nil {
		t.Fatal("expected invalid refresh error")
	}
}
