// SPDX-FileCopyrightText: 2022 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"testing"
)

func TestLatestDump(t *testing.T) {
	dir := filepath.Join("testdata", "dumps", "other", "pageview_complete")
	re := regexp.MustCompile(`^pageviews-(\d{8})-user\.bz2$`)
	got, err := LatestDump(dir, re)
	if err != nil {
		t.Error(err)
	}

	want := filepath.Join(dir, "2023", "2023-03", "pageviews-20230326-user.bz2")
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLatestDump_NoMatches(t *testing.T) {
	dir := filepath.Join("testdata", "dumps", "other", "pageview_complete")
	re := regexp.MustCompile(`^nosuchdump-(\d{8})\.bz2$`)
	if _, err := LatestDump(dir, re); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("got %v, want fs.ErrNotExist", err)
	}
}

func TestLatestDump_NoSuchDir(t *testing.T) {
	dir := filepath.Join("testdata", "dumps", "other", "no_such_dir")
	re := regexp.MustCompile(`^dump-(\d{8})\.bz2$`)
	if _, err := LatestDump(dir, re); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("got %v, want fs.ErrNotExist", err)
	}
}

func TestFormatLine(t *testing.T) {
	tests := []struct{ lang, site, title, value, expected string }{
		{"als", "wikipedia", "Wa\u0308he", "Q2595950",
			"gsw.wikipedia/wähe Q2595950"},
		{"az", "wikipedia", "Bakı", "Q9248",
			"az.wikipedia/bakı Q9248"},
		{"az", "wikipedia", "BAKI", "Q9248",
			"az.wikipedia/bakı Q9248"},
		{"azx", "wikipedia", "BAKI", "Q9248",
			"azx.wikipedia/baki Q9248"},
		{"bat_smg", "wikipedia", "Metā", "Q577",
			"sgs.wikipedia/metā Q577"},
		{"bat-smg", "wikipedia", "Metā", "Q577",
			"sgs.wikipedia/metā Q577"},
		{"be_x_old", "wikipedia", "Год", "Q577",
			"be-tarask.wikipedia/год Q577"},
		{"cbk_zam", "wikipedia", "Zamboanga Chavacano", "Q32174903",
			"cbk-x-zam.wikipedia/zamboanga_chavacano Q32174903"},
		{"cbk-zam", "wikipedia", "Zamboanga Chavacano", "Q32174903",
			"cbk-x-zam.wikipedia/zamboanga_chavacano Q32174903"},
		{"commons", "wikimedia", "Zwolle", "Q793",
			"und.commons/zwolle Q793"},
		{"de", "wikipedia", "BAKI", "Q9248",
			"de.wikipedia/baki Q9248"},
		{"de", "wikipedia", "Straße", "Q34442",
			"de.wikipedia/strasse Q34442"},
		{"fiu_vro", "wikipedia", "Aastak", "Q577",
			"vro.wikipedia/aastak Q577"},
		{"fiu-vro", "wikipedia", "Aastak", "Q577",
			"vro.wikipedia/aastak Q577"},
		{"incubator", "wikipedia", "Wp/cpx/Teng-cing-ch\u012b", "Q11736",
			"cpx.wikipedia/teng-cing-chī Q11736"},
		{"map_bms", "wikipedia", "Banyumasan", "Q33219",
			"jv-x-bms.wikipedia/banyumasan Q33219"},
		{"map-bms", "wikipedia", "Banyumasan", "Q33219",
			"jv-x-bms.wikipedia/banyumasan Q33219"},
		{"media", "mediawiki", "MediaWiki", "Q5296",
			"und.mediawiki/mediawiki Q5296"},
		{"meta", "wikimedia", "Main Page", "Q5296",
			"und.metawiki/main_page Q5296"},
		{"nds_nl", "wikipedia", "Zwolle", "Q793",
			"nds-NL.wikipedia/zwolle Q793"},
		{"nds-nl", "wikipedia", "Zwolle", "Q793",
			"nds-NL.wikipedia/zwolle Q793"},
		{"roa_rup", "wikipedia", "Anu", "Q577",
			"rup.wikipedia/anu Q577"},
		{"roa-rup", "wikipedia", "Anu", "Q577",
			"rup.wikipedia/anu Q577"},
		{"roa_tara", "wikipedia", "Àrvule", "Q10884",
			"nap-x-tara.wikipedia/àrvule Q10884"},
		{"roa-tara", "wikipedia", "Àrvule", "Q10884",
			"nap-x-tara.wikipedia/àrvule Q10884"},
		{"simple", "wikipedia", "Tianjin", "Q11736",
			"en-x-simple.wikipedia/tianjin Q11736"},
		{"sources", "wikipedia", "Author:蒋中正", "Q16574",
			"und.wikisource/author:蒋中正 Q16574"},
		{"species", "wiki", "Aepyceros melampus", "Q132576",
			"und.wikispecies/aepyceros_melampus Q132576"},
		{"tr", "wikipedia", "Diyarbakır", "Q83387",
			"tr.wikipedia/diyarbakır Q83387"},
		{"tr", "wikipedia", "DİYARBAKIR", "Q83387",
			"tr.wikipedia/diyarbakır Q83387"},
		{"xx", "wikipedia", "Space C", "U+0020",
			"xx.wikipedia/space_c U+0020"},
		{"xx", "wikipedia", "Tab\tC", "U+0007",
			"xx.wikipedia/tab_c U+0007"},
		{"xx", "wikipedia", "Zero\x00" + "C", "U+0000",
			"xx.wikipedia/zero_c U+0000"},
		{"zh_classical", "wikipedia", "尚書", "Q875313",
			"lzh.wikipedia/尚書 Q875313"},
		{"zh-classical", "wikipedia", "尚書", "Q875313",
			"lzh.wikipedia/尚書 Q875313"},
		{"zh_min_nan", "wikipedia", "Nî", "Q577",
			"nan.wikipedia/nî Q577"},
		{"zh-min-nan", "wikipedia", "Nî", "Q577",
			"nan.wikipedia/nî Q577"},
		{"zh_yue", "wikipedia", "\u5929\u6d25", "Q11736",
			"yue.wikipedia/天津 Q11736"},
		{"zh-yue", "wikipedia", "\u5929\u6d25", "Q11736",
			"yue.wikipedia/天津 Q11736"},
		{"", "commons", "Zwolle", "Q793",
			"und.commons/zwolle Q793"},
		{"", "wikidatawiki", "Project chat", "Q16503",
			"und.wikidata/project_chat Q16503"},
		{"", "wikimaniawiki", "Wikimania", "Q5296",
			"und.wikimania/wikimania Q5296"},
	}
	for _, c := range tests {
		if got := formatLine(c.lang, c.site, c.title, c.value); c.expected != got {
			msg := fmt.Sprintf("expected %q, got %q", c.expected, got)
			t.Error(msg)
		}
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct{ in, expected string }{
		{in: `"Foo:Bar"`, expected: "Foo:Bar"},
		{in: `"a\\a"`, expected: `a\a`},
		{in: `"a\/a"`, expected: "a/a"},
		{in: `"a\"a"`, expected: "a\"a"},
		{in: `"a'a"`, expected: "a'a"},
		{in: `"a\ba"`, expected: "a\ba"},
		{in: `"a\na"`, expected: "a\na"},
		{in: `"a\ra"`, expected: "a\ra"},
		{in: `"a\ta"`, expected: "a\ta"},
		{in: `"\uc11c\uacbd\uc8fc\uc5ed"`, expected: "서경주역"},
		{in: `"\u897f\u6176\u5dde\u99c5"`, expected: "西慶州駅"},
		{in: `"\u897f\u5e86\u5dde\u7ad9"`, expected: "西庆州站"},
	}
	for _, test := range tests {
		got, _ := unquote([]byte(test.in))
		if got != test.expected {
			t.Errorf("expected %q, got %q", test.expected, got)
		}
	}
}

func TestParseISOWeek(t *testing.T) {
	year, week, err := ParseISOWeek("2023-W07")
	if err != nil {
		t.Error(err)
	} else if year != 2023 || week != 7 {
		t.Errorf("want (2023, 7), got (%d, %d)", year, week)
	}
}

func TestParseISOWeek_BadFormat(t *testing.T) {
	if _, _, err := ParseISOWeek("2023-12-24"); err == nil {
		t.Error("want error, got nil")
	}
}

func ExampleParseISOWeek() {
	fmt.Println(ParseISOWeek("2018-W51")) // Output: 2018 51 <nil>
}

func TestISOWeekStart(t *testing.T) {
	for _, tc := range []struct {
		year     int
		day      int
		expected string
	}{
		{2018, -1, "2017-12-18"},
		{2018, 0, "2017-12-25"},
		{2018, 1, "2018-01-01"},
		{2018, 2, "2018-01-08"},
		{2019, 1, "2018-12-31"},
		{2019, 2, "2019-01-07"},
		{2019, 53, "2019-12-30"},
		{2019, 54, "2020-01-06"},
	} {
		got := ISOWeekStart(tc.year, tc.day).Format("2006-01-02")
		if tc.expected != got {
			t.Errorf("want ISOWeekStart(%d, %d) = %s, got %s", tc.year, tc.day, tc.expected, got)
		}
	}
}
