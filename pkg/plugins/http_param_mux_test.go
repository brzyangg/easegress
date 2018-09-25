package plugins

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	eghttp "github.com/hexdecteam/easegateway/pkg/http"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/hexdecteam/easegateway-types/plugins"
)

func newFastRequestHeaderByPath(path string, t *testing.T) plugins.Header {
	return newFastRequestHeader("http://localhost"+path, t)
}

func newFastRequestHeader(uri string, t *testing.T) plugins.Header {
	req := &fasthttp.Request{}
	req.SetRequestURI(uri)
	if u, err := url.Parse(uri); err != nil {
		t.Fatalf("newFastRequestHeader url.Parse failed, url: %s, erro: %v", uri, err)
	} else {
		req.Header.SetHost(u.Host)
	}

	return eghttp.NewFastRequestHeader(false, req.URI(), &req.Header)
}

func newNetRequestHeaderByPath(path string) plugins.Header {
	return newNetRequestHeader("http://localhost" + path)
}

func newNetRequestHeader(url string) plugins.Header {
	r, _ := http.NewRequest(http.MethodGet, url, nil /* body */)
	return eghttp.NewNetRequestHeader(r)
}

type parsePathTest struct {
	header         plugins.Header
	path           string
	pattern        string
	expectedErr    error
	expectedMatch  bool
	expectedParams map[string]string
}

var parsePathTests = []parsePathTest{
	{
		path:          "/r/megaease/easegateway/tags/",
		pattern:       "/r/{user}/{repo}/tags/",
		expectedMatch: true,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1",
		pattern:       "/r/{user}/{repo}/tags/",
		expectedMatch: false,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1",
		pattern:       "/r/{user}/{repo}/{tag}",
		expectedMatch: false,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "tags",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1",
		pattern:       "/r/{user}/{repo}/tags/{tag}",
		expectedMatch: true,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "server-0.1",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1/foo",
		pattern:       "/r/{user}/{repo}/tags/{tag}",
		expectedMatch: false,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "server-0.1",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1?foo=bar",
		pattern:       "/r/{user}/{repo}/tags/{tag}",
		expectedMatch: true,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "server-0.1"},
	},
	/*
		@zhiyan @longyun how about delete this test case ?
		current implementation parsePath() don't has url.query as path
		 {
				path:          "/r/megaease/easegateway/tags/server-0.1?foo=bar",
				pattern:       "/r/{user}/{repo}/tags/{tag}?{query}",
				expectedMatch: true,
				expectedParams: map[string]string{
					"user":  "megaease",
					"repo":  "easegateway",
					"tag":   "server-0.1",
					"query": "foo=bar",
				},
			},
	*/
	{
		path:          "/r/megaease/easegateway/tags/server-0.1/",
		pattern:       "/r/{user}/{repo}/tags/{tag}/{none}",
		expectedMatch: false,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "server-0.1",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1/foo",
		pattern:       "/r/{user}/{repo}/tags/{tag}/foo/{none}",
		expectedMatch: false,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "server-0.1",
		},
	},
	{
		path:          "/r/megaease/easegateway/tags/server-0.1/foo/bar",
		pattern:       "/r/{user}/{repo}/tags/{tag}/foo/{bar}",
		expectedMatch: true,
		expectedParams: map[string]string{
			"user": "megaease",
			"repo": "easegateway",
			"tag":  "server-0.1",
			"bar":  "bar",
		},
	},
	{
		path:           "/r/megaease",
		pattern:        "/r/megaease",
		expectedMatch:  true,
		expectedParams: map[string]string{},
	},
	{
		path:          "/{foo}/bar",
		pattern:       "/{foo}/{bar}",
		expectedMatch: true,
		expectedParams: map[string]string{
			"foo": "{foo}",
			"bar": "bar",
		},
	},
}

func testParsePathNormally(tests []parsePathTest, t *testing.T) {
	for i, test := range tests {
		match, params, err := parsePath(test.header.Path(), test.pattern)
		path := test.header.Path()
		pattern := test.pattern
		if err != test.expectedErr {
			t.Fatalf("#%d: \n expected err: %v, but got: %v", i, test.expectedErr, err)
		}
		if match != test.expectedMatch {
			t.Fatalf("\n#%d: \n path: %s\n pattern: %s\n expected match: %v, but got: %v",
				i, path, pattern, test.expectedMatch, match)
		}
		if err == nil /* && match */ && !reflect.DeepEqual(params, test.expectedParams) {
			t.Fatalf("\n#%d: \n path: %s\n pattern: %s\n expected params: %v, but got: %v", i, path, pattern, test.expectedParams, params)
		}
	}
}

func TestFastHTTPParsePath(t *testing.T) {
	for i := range parsePathTests {
		parsePathTests[i].header = newFastRequestHeaderByPath(parsePathTests[i].path, t)
	}
	testParsePathNormally(parsePathTests, t)
}

func TestNetHTTPParsePath(t *testing.T) {
	for i := range parsePathTests {
		parsePathTests[i].header = newNetRequestHeaderByPath(parsePathTests[i].path)
	}
	testParsePathNormally(parsePathTests, t)
}

func TestParsePathExceptionally(t *testing.T) {
	path := "/r/megaease"
	pattern := "/r/{user"

	match, ret, err := parsePath(path, pattern)
	if err == nil {
		t.Fatalf("expected error unraied %v, %v", match, ret)
	}
}

func TestDuplicatedPathNormally(t *testing.T) {
	path1 := "/r/abc"
	path2 := "/r/def"

	dup, err := duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc"
	path2 = "/r/abc/def"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/"
	path2 = "/r/abc/def"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/defg"
	path2 = "/r/abc/def"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/defg"
	path2 = "/r/abc/def/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/{foo}"
	path2 = "/r/abc/def/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/{foo}/"
	path2 = "/r/abc/def"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/{foo}/"
	path2 = "/r/abc/def/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/def/"
	path2 = "/r/abc/{foo}/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/r/abc/{foo}/{none}"
	path2 = "/r/abc/def/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/test.html"
	path2 = "/{page}"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/{page}"
	path2 = "/test.html"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/{foo}/bar"
	path2 = "/foo/{bar}"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/{foo}/bar/"
	path2 = "/foo/{bar}"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/{foo}/bar"
	path2 = "/foo/{bar}/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/{foo}/bar/baz"
	path2 = "/foo/{bar}/baz"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !dup {
		t.Fatalf("unexpected check return %v", dup)
	}

	path1 = "/{foo}/bar/{baz}"
	path2 = "/foo/{bar}/"

	dup, err = duplicatedPath(path1, path2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if dup {
		t.Fatalf("unexpected check return %v", dup)
	}
}

////

func TestParamMuxDifferentGenerations(t *testing.T) {
	m := mustNewParamMux(t)
	pluginA := "plugin-a"
	pluginA1 := mockHTTPInput(pluginA)
	pluginA2 := mockHTTPInput(pluginA)
	ctx1 := mockPipelineContext("pipeline-1", []string{pluginA})
	entryA1_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA1)
	entryA1_2 := mockHTTPMuxEntry("", "", "", "/a", "", "", "POST", 0, pluginA1)
	entryA2_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA2)
	entryA2_2 := mockHTTPMuxEntry("", "", "", "/a", "", "", "POST", 0, pluginA2)
	entryA2_3 := mockHTTPMuxEntry("", "", "", "/a", "", "", "PUT", 0, pluginA2)

	mustAddFunc(t, m, ctx1, entryA1_1)
	mustAddFunc(t, m, ctx1, entryA1_2)
	mustAddFunc(t, m, ctx1, entryA2_1)
	resultRTable := map[string]map[string]map[string]*plugins.HTTPMuxEntry{
		ctx1.PipelineName(): {
			"/a": {
				"GET": entryA2_1,
			},
		},
	}
	mustGetParamMuxResult(t, m, resultRTable)

	mustAddFunc(t, m, ctx1, entryA2_2)
	mustAddFunc(t, m, ctx1, entryA2_3)
	resultRTable = map[string]map[string]map[string]*plugins.HTTPMuxEntry{
		ctx1.PipelineName(): {
			"/a": {
				"GET":  entryA2_1,
				"POST": entryA2_2,
				"PUT":  entryA2_3,
			},
		},
	}
	mustGetParamMuxResult(t, m, resultRTable)
}

func TestParamMuxCleanOutdatedEntries(t *testing.T) {
	m := mustNewParamMux(t)
	pluginA := "plugin-a"
	pluginA1 := mockHTTPInput(pluginA)
	pluginA2 := mockHTTPInput(pluginA)
	ctx1 := mockPipelineContext("pipeline-1", []string{pluginA})
	ctx2 := mockPipelineContext("pipeline-2", []string{pluginA})
	entryA1_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA1)
	entryA1_2 := mockHTTPMuxEntry("", "", "", "/a", "", "", "POST", 0, pluginA1)
	entryA2_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA2)

	mustAddFunc(t, m, ctx1, entryA1_1)
	mustAddFunc(t, m, ctx1, entryA1_2)
	pipelineEntries := mustDeleteFuncs(t, m, ctx1)
	// NOTICE: Even the plugin comes back from a different pipeline, the older rules
	// of it (same plugin name) will be cleaned too.
	mustAddFunc(t, m, ctx2, entryA2_1)
	mustAddFuncs(t, m, ctx2, pipelineEntries)

	resultRTable := map[string]map[string]map[string]*plugins.HTTPMuxEntry{
		ctx2.PipelineName(): {
			"/a": {
				"GET": entryA2_1,
			},
		},
	}
	mustGetParamMuxResult(t, m, resultRTable)
}

func TestParamMuxCleanDeadEntries(t *testing.T) {
	m := mustNewParamMux(t)
	pluginA := "plugin-a"
	pluginB := "plugin-b"
	pluginA1 := mockHTTPInput(pluginA)
	pluginB1 := mockHTTPInput(pluginB)
	ctx1_1 := mockPipelineContext("pipeline-1", []string{pluginA, pluginB})
	ctx1_2 := mockPipelineContext("pipeline-1", []string{pluginA})
	entryA1_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA1)
	entryB1_1 := mockHTTPMuxEntry("", "", "", "/b", "", "", "GET", 0, pluginB1)
	entryB1_2 := mockHTTPMuxEntry("", "", "", "/b", "", "", "POST", 0, pluginB1)

	mustAddFunc(t, m, ctx1_1, entryA1_1)
	mustAddFunc(t, m, ctx1_1, entryB1_1)
	mustAddFunc(t, m, ctx1_1, entryB1_2)
	pipelineEntries := mustDeleteFuncs(t, m, ctx1_1)
	// NOTICE: The absence of pluginB in ctx2 leads to clean all entryB*.
	m.AddFuncs(ctx1_2, pipelineEntries)

	resultRTable := map[string]map[string]map[string]*plugins.HTTPMuxEntry{
		ctx1_2.PipelineName(): {
			"/a": {
				"GET": entryA1_1,
			},
		},
	}
	mustGetParamMuxResult(t, m, resultRTable)
}

func TestParamMuxFatigue(t *testing.T) {
	m := mustNewParamMux(t)
	pluginA := "plugin-a"
	pluginB := "plugin-b"
	pluginC := "plugin-c"
	pluginD := "plugin-d"
	pluginA1 := mockHTTPInput(pluginA)
	pluginA2 := mockHTTPInput(pluginA)
	pluginB1 := mockHTTPInput(pluginB)
	pluginB2 := mockHTTPInput(pluginB)
	pluginC1 := mockHTTPInput(pluginC)
	pluginC2 := mockHTTPInput(pluginC)
	pluginC3 := mockHTTPInput(pluginC)
	pluginD1 := mockHTTPInput(pluginD)
	pluginD2 := mockHTTPInput(pluginD)
	pluginD3 := mockHTTPInput(pluginD)
	pluginD4 := mockHTTPInput(pluginD)

	ctx1 := mockPipelineContext("pipeline-1", []string{pluginA, pluginB, pluginC})
	ctx2 := mockPipelineContext("pipeline-2", []string{pluginD})
	// add entry
	entryA1_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA1)
	entryA1_2 := mockHTTPMuxEntry("", "", "", "/a", "", "", "POST", 0, pluginA1)
	entryA1_3 := mockHTTPMuxEntry("", "", "", "/a", "", "", "PUT", 0, pluginA1)
	entryA2_1 := mockHTTPMuxEntry("", "", "", "/a", "", "", "GET", 0, pluginA2)
	entryA2_2 := mockHTTPMuxEntry("", "", "", "/a", "", "", "POST", 0, pluginA2)
	messA := func() {
		mustAddFunc(t, m, ctx1, entryA1_1)
		mustAddFunc(t, m, ctx1, entryA1_2)
		mustAddFunc(t, m, ctx1, entryA1_3)
		mustAddFunc(t, m, ctx1, entryA2_1)
		mustDeleteFunc(t, m, ctx1, entryA1_1)
		mustDeleteFunc(t, m, ctx1, entryA1_2)
		mustAddFunc(t, m, ctx1, entryA2_2)
		mustDeleteFunc(t, m, ctx1, entryA1_3)
	}

	// delete entry
	entryB1_1 := mockHTTPMuxEntry("", "", "", "/b", "", "", "GET", 1, pluginB1)
	entryB1_2 := mockHTTPMuxEntry("", "", "", "/b", "", "", "DELETE", 1, pluginB1)
	entryB2_1 := mockHTTPMuxEntry("", "", "", "/b", "", "", "GET", 1, pluginB2)
	messB := func() {
		mustAddFunc(t, m, ctx1, entryB1_1)
		mustAddFunc(t, m, ctx1, entryB1_2)
		mustAddFunc(t, m, ctx1, entryB2_1)
		mustDeleteFunc(t, m, ctx1, entryB1_1)
		// mock missing mustDeleteFunc(t, m, ctx1, entryB1_1)
	}

	// no change
	entryC1_1 := mockHTTPMuxEntry("", "", "", "/c", "", "", "GET", 1, pluginC1)
	entryC1_2 := mockHTTPMuxEntry("", "", "", "/c", "", "", "HEAD", 1, pluginC1)
	entryC2_1 := mockHTTPMuxEntry("", "", "", "/c", "", "", "GET", 1, pluginC2)
	entryC2_2 := mockHTTPMuxEntry("", "", "", "/c", "", "", "HEAD", 1, pluginC2)
	entryC3_1 := mockHTTPMuxEntry("", "", "", "/c", "", "", "GET", 1, pluginC3)
	entryC3_2 := mockHTTPMuxEntry("", "", "", "/c", "", "", "HEAD", 1, pluginC3)
	messC := func() {
		mustAddFunc(t, m, ctx1, entryC1_1)
		mustAddFunc(t, m, ctx1, entryC1_2)
		pipelineEntries1 := mustDeleteFuncs(t, m, ctx1)
		mustDeleteFunc(t, m, ctx1, entryC1_1)
		mustAddFunc(t, m, ctx1, entryC2_1)
		mustAddFuncs(t, m, ctx1, pipelineEntries1)
		// mock missing mustDeleteFunc(t, m, ctx1, entryC1_2)
		mustAddFunc(t, m, ctx1, entryC2_2)
		mustAddFunc(t, m, ctx1, entryC3_1)
		mustDeleteFunc(t, m, ctx1, entryC2_1)
		mustAddFunc(t, m, ctx1, entryC3_2)
		mustDeleteFunc(t, m, ctx1, entryC2_2)
	}

	// mess up
	entryD1_1 := mockHTTPMuxEntry("", "", "", "/d", "", "", "GET", 2, pluginD1)
	entryD2_1 := mockHTTPMuxEntry("", "", "", "/dd", "", "", "POST", 20, pluginD2)
	entryD3_1 := mockHTTPMuxEntry("", "", "", "/ddd", "", "", "PUT", 200, pluginD3)
	entryD4_1 := mockHTTPMuxEntry("", "", "", "/dddd", "", "", "GET", 2000, pluginD4)
	entryD4_2 := mockHTTPMuxEntry("", "", "", "/dddd", "", "", "POST", 2000, pluginD4)
	entryD4_3 := mockHTTPMuxEntry("", "", "", "/dddd", "", "", "PUT", 2000, pluginD4)
	messD := func() {
		mustAddFunc(t, m, ctx2, entryD1_1)
		mustAddFunc(t, m, ctx2, entryD2_1)
		// mock missing mustDeleteFunc(t, m, ctx2, entryD1_1)
		mustAddFunc(t, m, ctx2, entryD3_1)
		mustDeleteFunc(t, m, ctx2, entryD3_1)
		mustDeleteFunc(t, m, ctx2, entryD2_1)
		mustAddFunc(t, m, ctx2, entryD4_1)
		mustAddFunc(t, m, ctx2, entryD4_2)
		mustAddFunc(t, m, ctx2, entryD4_3)
	}

	messA()
	messB()
	messC()
	messD()

	resultRTable := map[string]map[string]map[string]*plugins.HTTPMuxEntry{
		ctx1.PipelineName(): {
			"/a": {
				"GET":  entryA2_1,
				"POST": entryA2_2,
			},
			"/b": {
				"GET": entryB2_1,
			},
			"/c": {
				"GET":  entryC3_1,
				"HEAD": entryC3_2,
			},
		},
		ctx2.PipelineName(): {
			"/dddd": {
				"GET":  entryD4_1,
				"POST": entryD4_2,
				"PUT":  entryD4_3,
			},
		},
	}
	mustGetParamMuxResult(t, m, resultRTable)

	////

	pipelineEntries := mustDeleteFuncs(t, m, ctx1)
	// delete pluginA pluginC
	ctx1 = mockPipelineContext("pipeline-1", []string{pluginB})
	mustAddFuncs(t, m, ctx1, pipelineEntries)

	resultRTable = map[string]map[string]map[string]*plugins.HTTPMuxEntry{
		ctx1.PipelineName(): {
			"/b": {
				"GET": entryB2_1,
			},
		},
		ctx2.PipelineName(): {
			"/dddd": {
				"GET":  entryD4_1,
				"POST": entryD4_2,
				"PUT":  entryD4_3,
			},
		},
	}
	mustGetParamMuxResult(t, m, resultRTable)
}