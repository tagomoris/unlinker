package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertEqualStrings(t *testing.T, a []string, b []string, message string) {
	if len(a) == len(b) {
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("Inequal at [%v]: %v != %v", i, a, b)
			}
		}
	} else {
		t.Fatalf("Inequal number of members: %v != %v", a, b)
	}
}

func assertFileExists(t *testing.T, path string) {
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().IsRegular() {
		return
	}
	t.Fatalf("Path[%v] is not a regular file", path)
}

func assertFileMissing(t *testing.T, path string) {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Fatalf("Path[%v] should not exist, but exists", path)
}

func TestGlobMatchPart(t *testing.T) {
	assertEqual(t, "1", globMatchPart("/my/path/access_log.1", "/my/path/access_log.*"), "")
	assertEqual(t, "20180617", globMatchPart("/my/path/access_log.20180617.log", "/my/path/access_log.*.log"), "")
}

func TestConvertTimestampFormat(t *testing.T) {
	// Golang time format template: Mon Jan 2 15:04:05 -0700 MST 2006
	var format string
	var ct bool

	format, ct = convertTimestampFormat("%Y-%m-%d_%H%M%S")
	assertEqual(t, false, ct, "")
	assertEqual(t, "2006-01-02_150405", format, "")

	format, ct = convertTimestampFormat("%Y%m%d%H%%%Z")
	assertEqual(t, true, ct, "")
	assertEqual(t, "2006010215%MST", format, "")
}

func TestFindTargetsWithAge(t *testing.T) {
	paths := []string{"/a/b/c.1", "/a/b/c.2", "/a/b/c.3", "/a/b/c.4", "/a/b/c.10"}
	expected := []string{"/a/b/c.3", "/a/b/c.4", "/a/b/c.10"}
	assertEqualStrings(t, expected, findTargetsWithAge("/a/b/c.*", paths, 3), "")
}

func TestFindTargetsWithTimestamp(t *testing.T) {
	paths := []string{"/a/b/c.20180617_23", "/a/b/c.20180618_00", "/a/b/c.20180618_01", "/a/b/c.20180618_02", "/a/b/c.20180618_03", "/a/b/c.20180618_04", "/a/b/c.20180618_05", "/a/b/c.20180618_10"}
	expected := []string{"/a/b/c.20180617_23", "/a/b/c.20180618_00"}
	now := time.Date(2018, 6, 18, 12, 0, 0, 0, time.Local) // 2018-06-18 12:00:00.0 localtime
	assertEqualStrings(t, expected, findTargetsWithTimestamp(now, "/a/b/c.*", paths, "%Y%m%d_%H", 86400 / 2), "")

 	jst := time.FixedZone("Japanese Standard Time", int((9 * time.Hour).Seconds()))

	paths = []string{"/a/b/c.2018061723_+0900.log.gz", "/a/b/c.2018061800_+0900.log.gz", "/a/b/c.2018061801_+0900.log.gz", "/a/b/c.2018061802_+0900.log.gz", "/a/b/c.2018061803_+0900.log.gz", "/a/b/c.2018061804_+0900.log.gz", "/a/b/c.2018061805_+0900.log.gz", "/a/b/c.2018061810_+0900.log.gz"}
	expected = []string{"/a/b/c.2018061723_+0900.log.gz", "/a/b/c.2018061800_+0900.log.gz"}
	now = time.Date(2018, 6, 18, 12, 0, 0, 0, jst) // 2018-06-18 12:00:00.0
	assertEqualStrings(t, expected, findTargetsWithTimestamp(now, "/a/b/c.*.log.gz", paths, "%Y%m%d%H_%z", 86400 / 2), "")
}

func prepareTempDir() string {
	dir, err := ioutil.TempDir("", "unlinker")
	if err != nil {
		panic("failed to create tempdir")
	}
	return dir
}

func prepareTestFile(dir, basename string, mtime time.Time) string {
	fullpath := filepath.Join(dir, basename)
	if err := ioutil.WriteFile(fullpath, []byte{}, 0644); err != nil {
		panic("failed to create a file:" + fullpath)
	}
	if err := os.Chtimes(fullpath, mtime, mtime); err != nil {
		panic("failed to modify mtime:" + fullpath)
	}
	return fullpath
}

func TestFindTargetsWithMtime(t *testing.T) {
	tmpdir := prepareTempDir()
	names := []string{"c.20180617_23", "c.20180618_00", "c.20180618_01", "c.20180618_02", "c.20180618_03", "c.20180618_04", "c.20180618_05", "c.20180618_10"}
	paths := []string{
		prepareTestFile(tmpdir, names[0], time.Date(2018, 6, 18, 0, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[1], time.Date(2018, 6, 18, 1, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[2], time.Date(2018, 6, 18, 2, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[3], time.Date(2018, 6, 18, 3, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[4], time.Date(2018, 6, 18, 4, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[5], time.Date(2018, 6, 18, 5, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[6], time.Date(2018, 6, 18, 6, 0, 3, 0, time.Local)),
		prepareTestFile(tmpdir, names[7], time.Date(2018, 6, 18, 11, 0, 3, 0, time.Local)),
	}

	now := time.Date(2018, 6, 18, 12, 0, 0, 0, time.Local) // 2018-06-18 12:00:00.0 localtime
	expected := []string{}
	assertEqualStrings(t, expected, findTargetsWithMtime(now, paths, 86400 / 2), "")

	now = time.Date(2018, 6, 18, 13, 0, 0, 0, time.Local) // 2018-06-18 13:00:00.0 localtime
	expected = []string{paths[0]}
	assertEqualStrings(t, expected, findTargetsWithMtime(now, paths, 86400 / 2), "")

	now = time.Date(2018, 6, 18, 16, 0, 0, 0, time.Local) // 2018-06-18 13:00:00.0 localtime
	expected = []string{paths[0], paths[1], paths[2], paths[3]}
	assertEqualStrings(t, expected, findTargetsWithMtime(now, paths, 86400 / 2), "")
}

func TestRun(t *testing.T) {
	tmpdir := prepareTempDir()
	confDir, err := ioutil.TempDir(tmpdir, "conf")
	if err != nil {
		panic("failed to create confDir")
	}

	dataDir1, err := ioutil.TempDir(tmpdir, "log1")
	if err != nil {
		panic("failed to create dataDir1")
	}
	conf1 := fmt.Sprintf("{\"rule\":\"age\",\"path\":\"%s\",\"age\":3}", filepath.Join(dataDir1, "c.*"))
	ioutil.WriteFile(filepath.Join(confDir, "age.json"), []byte(conf1), 0644)

	dataDir2, err := ioutil.TempDir(tmpdir, "log2")
	if err != nil {
		panic("failed to create dataDir2")
	}
	conf2 := fmt.Sprintf("{\"rule\":\"timestamp\",\"path\":\"%s\",\"format\":\"%s\",\"expire_sec\":%d}", filepath.Join(dataDir2, "c.*.log.gz"), "%Y%m%d%H", 86400 / 2)
	ioutil.WriteFile(filepath.Join(confDir, "timestamp.json"), []byte(conf2), 0644)

	dataDir3, err := ioutil.TempDir(tmpdir, "log3")
	if err != nil {
		panic("failed to create dataDir3")
	}
	conf3 := fmt.Sprintf("{\"rule\":\"mtime\",\"path\":\"%s\",\"expire_sec\":%d}", filepath.Join(dataDir3, "c.*"), 86400 / 2)
	ioutil.WriteFile(filepath.Join(confDir, "mtime.json"), []byte(conf3), 0644)

	now := time.Date(2018, 6, 18, 16, 0, 0, 0, time.Local) // 2018-06-18 16:00:00.0 localtime

	ageNames := []string{"c", "c.1", "c.2", "c.3", "c.4", "c.5", "c.10"}
	agePaths := []string{
		prepareTestFile(dataDir1, ageNames[0], time.Date(2018, 6, 18, 13, 3, 50, 0, time.Local)),
		prepareTestFile(dataDir1, ageNames[1], time.Date(2018, 6, 18, 12, 3, 50, 0, time.Local)),
		prepareTestFile(dataDir1, ageNames[2], time.Date(2018, 6, 18, 11, 3, 50, 0, time.Local)),
		prepareTestFile(dataDir1, ageNames[3], time.Date(2018, 6, 18, 10, 3, 50, 0, time.Local)),
		prepareTestFile(dataDir1, ageNames[4], time.Date(2018, 6, 18, 9, 3, 50, 0, time.Local)),
		prepareTestFile(dataDir1, ageNames[5], time.Date(2018, 6, 18, 8, 3, 50, 0, time.Local)),
		prepareTestFile(dataDir1, ageNames[6], time.Date(2018, 6, 18, 3, 3, 50, 0, time.Local)),
	}
	timestampNames := []string{"c.2018061723.log.gz", "c.2018061800.log.gz", "c.2018061801.log.gz", "c.2018061802.log.gz", "c.2018061803.log.gz", "c.2018061804.log.gz", "c.2018061805.log.gz", "c.2018061810.log.gz"}
	timestampPaths := []string{
		prepareTestFile(dataDir2, timestampNames[0], time.Date(2018, 6, 18, 0, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir2, timestampNames[1], time.Date(2018, 6, 18, 1, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir2, timestampNames[2], time.Date(2018, 6, 18, 2, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir2, timestampNames[3], time.Date(2018, 6, 18, 3, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir2, timestampNames[4], time.Date(2018, 6, 18, 4, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir2, timestampNames[5], time.Date(2018, 6, 18, 5, 0, 1, 0, time.Local)), // 12hours ago (by timestamp in path)
		prepareTestFile(dataDir2, timestampNames[6], time.Date(2018, 6, 18, 6, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir2, timestampNames[7], time.Date(2018, 6, 18, 11, 0, 1, 0, time.Local)),
	}
	mtimeNames := []string{"c.20180617_23", "c.20180618_00", "c.20180618_01", "c.20180618_02", "c.20180618_03", "c.20180618_04", "c.20180618_05", "c.20180618_10"}
	mtimePaths := []string{
		prepareTestFile(dataDir3, mtimeNames[0], time.Date(2018, 6, 18, 0, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir3, mtimeNames[1], time.Date(2018, 6, 18, 1, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir3, mtimeNames[2], time.Date(2018, 6, 18, 2, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir3, mtimeNames[3], time.Date(2018, 6, 18, 3, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir3, mtimeNames[4], time.Date(2018, 6, 18, 4, 0, 1, 0, time.Local)), // 12hours - 1s ago (by mtime)
		prepareTestFile(dataDir3, mtimeNames[5], time.Date(2018, 6, 18, 5, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir3, mtimeNames[6], time.Date(2018, 6, 18, 6, 0, 1, 0, time.Local)),
		prepareTestFile(dataDir3, mtimeNames[7], time.Date(2018, 6, 18, 11, 0, 1, 0, time.Local)),
	}

	run(confDir, now)

	assertFileExists(t, agePaths[0])
	assertFileExists(t, agePaths[1])
	assertFileExists(t, agePaths[2])
	assertFileMissing(t, agePaths[3])
	assertFileMissing(t, agePaths[4])
	assertFileMissing(t, agePaths[5])
	assertFileMissing(t, agePaths[6])

	assertFileMissing(t, timestampPaths[0])
	assertFileMissing(t, timestampPaths[1])
	assertFileMissing(t, timestampPaths[2])
	assertFileMissing(t, timestampPaths[3])
	assertFileMissing(t, timestampPaths[4])
	assertFileMissing(t, timestampPaths[5])
	assertFileExists(t, timestampPaths[6])
	assertFileExists(t, timestampPaths[7])

	assertFileMissing(t, mtimePaths[0])
	assertFileMissing(t, mtimePaths[1])
	assertFileMissing(t, mtimePaths[2])
	assertFileMissing(t, mtimePaths[3])
	assertFileExists(t, mtimePaths[4])
	assertFileExists(t, mtimePaths[5])
	assertFileExists(t, mtimePaths[6])
	assertFileExists(t, mtimePaths[7])
}
