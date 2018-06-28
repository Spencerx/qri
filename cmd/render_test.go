package cmd

import (
	"testing"

	"github.com/qri-io/cafs"
	"github.com/qri-io/qri/lib"
)

func TestRenderComplete(t *testing.T) {
	streams, in, out, errs := NewTestIOStreams()
	setNoColor(true)

	f, err := NewTestFactory()
	if err != nil {
		t.Errorf("error creating new test factory: %s", err)
		return
	}

	cases := []struct {
		args   []string
		expect string
		err    string
	}{
		{[]string{}, "", ""},
		{[]string{"test"}, "test", ""},
		{[]string{"test", "test2"}, "test", ""},
	}

	for i, c := range cases {
		opt := &RenderOptions{
			IOStreams: streams,
		}

		opt.Complete(f, c.args)

		if c.err != errs.String() {
			t.Errorf("case %d, error mismatch. Expected: '%s', Got: '%s'", i, c.err, errs.String())
			ioReset(in, out, errs)
			continue
		}

		if c.expect != opt.Ref {
			t.Errorf("case %d, opt.Ref not set correctly. Expected: '%s', Got: '%s'", i, c.expect, opt.Ref)
			ioReset(in, out, errs)
			continue
		}

		if opt.RenderRequests == nil {
			t.Errorf("case %d, opt.RenderRequests not set.", i)
			ioReset(in, out, errs)
			continue
		}
		ioReset(in, out, errs)
	}
}

func TestRenderValidate(t *testing.T) {
	cases := []struct {
		ref string
		err string
		msg string
	}{
		{"", ErrBadArgs.Error(), "peername and dataset name needed in order to render, for example:\n   $ qri render me/dataset_name\nsee `qri render --help` from more info"},
		{"me/test", "", ""},
	}
	for i, c := range cases {
		opt := &RenderOptions{
			Ref: c.ref,
		}

		err := opt.Validate()
		if (err == nil && c.err != "") || (err != nil && c.err != err.Error()) {
			t.Errorf("case %d, mismatched error. Expected: %s, Got: %s", i, c.err, err)
			continue
		}
		if libErr, ok := err.(lib.Error); ok {
			if libErr.Message() != c.msg {
				t.Errorf("case %d, mismatched user-friendly message. Expected: '%s', Got: '%s'", i, c.msg, libErr.Message())
				continue
			}
		} else if c.msg != "" {
			t.Errorf("case %d, mismatched user-friendly message. Expected: '%s', Got: ''", i, c.msg)
			continue
		}
	}
}

func TestRenderRun(t *testing.T) {
	streams, in, out, errs := NewTestIOStreams()
	setNoColor(true)

	f, err := NewTestFactory()
	if err != nil {
		t.Errorf("error creating new test factory: %s", err)
		return
	}

	templateFile := cafs.NewMemfileBytes("template.html", []byte(`<html><h2>{{.Peername}}/{{.Name}}</h2></html>`))

	repo, err := f.Repo()
	if err != nil {
		t.Errorf("error getting repo from factory: %s", err)
		return
	}

	key, err := repo.Store().Put(templateFile, false)
	if err != nil {
		t.Errorf("error putting template into store: %s", err)
		return
	}

	cfg, err := f.Config()
	if err != nil {
		t.Errorf("error getting config from factory: %s", err)
		return
	}

	if err := cfg.Set("render.defaultTemplateHash", key.String()); err != nil {
		t.Errorf("error setting default template in config: %s", err)
		return
	}
	lib.Config = cfg

	cases := []struct {
		ref      string
		template string
		output   string
		all      bool
		limit    int
		offset   int
		expected string
		err      string
		msg      string
	}{
		{"peer/bad_dataset", "", "", false, 10, 0, "", "repo: not found", "could not find dataset 'peer/bad_dataset'"},
		{"peer/cities", "", "", false, 10, 0, "<html><h2>peer/cities</h2></html>", "", ""},
		{"peer/cities", "testdata/template.html", "", false, 2, 0, "<html><h2>peer/cities</h2><tbody><tr><td>toronto</td><td>40000000</td><td>55.5</td><td>false</td></tr><tr><td>new york</td><td>8500000</td><td>44.4</td><td>true</td></tr></tbody></html>", "", ""},
		{"peer/cities", "testdata/template.html", "", false, 1, 2, "<html><h2>peer/cities</h2><tbody><tr><td>chicago</td><td>300000</td><td>44.4</td><td>true</td></tr></tbody></html>", "", ""},
	}

	for i, c := range cases {
		rr, err := f.RenderRequests()
		if err != nil {
			t.Errorf("case %d, error creating dataset request: %s", i, err)
			continue
		}

		opt := &RenderOptions{
			IOStreams:      streams,
			Ref:            c.ref,
			Template:       c.template,
			Output:         c.output,
			All:            c.all,
			Limit:          c.limit,
			Offset:         c.offset,
			RenderRequests: rr,
		}

		err = opt.Run()
		if (err == nil && c.err != "") || (err != nil && c.err != err.Error()) {
			t.Errorf("case %d, mismatched error. Expected: '%s', Got: '%v'", i, c.err, err)
			ioReset(in, out, errs)
			continue
		}

		if libErr, ok := err.(lib.Error); ok {
			if libErr.Message() != c.msg {
				t.Errorf("case %d, mismatched user-friendly message. Expected: '%s', Got: '%s'", i, c.msg, libErr.Message())
				ioReset(in, out, errs)
				continue
			}
		} else if c.msg != "" {
			t.Errorf("case %d, mismatched user-friendly message. Expected: '%s', Got: ''", i, c.msg)
			ioReset(in, out, errs)
			continue
		}

		if c.expected != out.String() {
			t.Errorf("case %d, output mismatch. Expected: '%s', Got: '%s'", i, c.expected, out.String())
			ioReset(in, out, errs)
			continue
		}
		ioReset(in, out, errs)
	}
}
