package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/cybozu-go/sabakan/v2"
	ign22 "github.com/flatcar/ignition/config/v2_2/types"
	ign23 "github.com/flatcar/ignition/config/v2_3/types"
	"github.com/vincent-petithory/dataurl"
)

type renderFunc func(name, tmpl string) (string, error)

func (s Server) handleIgnitions(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/boot/ignitions/"):], "/")
	if len(params) != 2 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	serial := params[0]
	id := params[1]

	if len(serial) == 0 || len(id) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	if r.Method != "GET" {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	m, err := s.Model.Machine.Get(r.Context(), serial)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	tmpl, err := s.Model.Ignition.GetTemplate(r.Context(), m.Spec.Role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	ign, err := s.renderIgnition(tmpl, m)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, ign, http.StatusOK)
}

func (s Server) renderIgnition(tmpl *sabakan.IgnitionTemplate, m *sabakan.Machine) (interface{}, error) {
	myURL := s.MyURL.String()
	tmplFuncs := template.FuncMap{
		"MyURL": func() string { return myURL },
		"Metadata": func(key string) (interface{}, error) {
			val, ok := tmpl.Metadata[key]
			if !ok {
				return nil, errors.New("no such meta data: " + key)
			}
			return val, nil
		},
		"json": jsonFunc,
		"add":  addFunc,
		"sub":  subFunc,
		"mul":  mulFunc,
		"div":  divFunc,
	}
	render := func(name, tmpl string) (string, error) {
		buf := &bytes.Buffer{}
		t, err := template.New(name).Funcs(tmplFuncs).Parse(tmpl)
		if err != nil {
			return "", err
		}
		err = t.Execute(buf, m)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	switch tmpl.Version {
	case sabakan.Ignition2_2:
		return renderIgnition2_2(tmpl, render)
	case sabakan.Ignition2_3:
		return renderIgnition2_3(tmpl, render)
	}

	return nil, errors.New("unsupported ignition version: " + string(tmpl.Version))
}

func renderIgnition2_2(tmpl *sabakan.IgnitionTemplate, render renderFunc) (interface{}, error) {
	ign := new(ign22.Config)
	err := json.Unmarshal([]byte(tmpl.Template), ign)
	if err != nil {
		return nil, err
	}

	ign.Ignition.Version = "2.2.0"
	for i := range ign.Passwd.Groups {
		g := &ign.Passwd.Groups[i]
		pfx := fmt.Sprintf("passwd.groups[%d].", i)
		g.Name, err = render(pfx+"name", g.Name)
		if err != nil {
			return nil, err
		}
		g.PasswordHash, err = render(pfx+"passwordHash", g.PasswordHash)
		if err != nil {
			return nil, err
		}
	}
	for i := range ign.Passwd.Users {
		u := &ign.Passwd.Users[i]
		pfx := fmt.Sprintf("passwd.users[%d].", i)
		u.Name, err = render(pfx+"name", u.Name)
		if err != nil {
			return nil, err
		}
		if u.PasswordHash != nil {
			passwd, err := render(pfx+"passwordHash", *u.PasswordHash)
			if err != nil {
				return nil, err
			}
			u.PasswordHash = &passwd
		}
		for j, key := range u.SSHAuthorizedKeys {
			name := pfx + fmt.Sprintf("sshAuthorizedKeys[%d]", j)
			k, err := render(name, string(key))
			if err != nil {
				return nil, err
			}
			u.SSHAuthorizedKeys[j] = ign22.SSHAuthorizedKey(k)
		}
		u.Gecos, err = render(pfx+"gecos", u.Gecos)
		if err != nil {
			return nil, err
		}
		u.HomeDir, err = render(pfx+"homeDir", u.HomeDir)
		if err != nil {
			return nil, err
		}
		u.PrimaryGroup, err = render(pfx+"primaryGroup", u.PrimaryGroup)
		if err != nil {
			return nil, err
		}
		for j, group := range u.Groups {
			name := pfx + fmt.Sprintf("groups[%d]", j)
			g, err := render(name, string(group))
			if err != nil {
				return nil, err
			}
			u.Groups[j] = ign22.Group(g)
		}
		u.Shell, err = render(pfx+"shell", u.Shell)
		if err != nil {
			return nil, err
		}
	}

	for i := range ign.Storage.Files {
		file := &ign.Storage.Files[i]
		if file.Contents.Verification.Hash != nil {
			h, err := render(file.Path, *file.Contents.Verification.Hash)
			if err != nil {
				return nil, err
			}
			file.Contents.Verification.Hash = &h
		}
		if !strings.HasPrefix(file.Contents.Source, "data:") {
			file.Contents.Source, err = render(file.Path, file.Contents.Source)
			if err != nil {
				return nil, err
			}
			continue
		}

		// Render the file contents if embedded.
		d, err := dataurl.DecodeString(file.Contents.Source)
		if err != nil {
			return nil, err
		}
		rendered, err := render(file.Path, string(d.Data))
		if err != nil {
			return nil, err
		}
		file.Contents.Source = "data:," + dataurl.EscapeString(rendered)
	}

	for i := range ign.Networkd.Units {
		unit := &ign.Networkd.Units[i]
		unit.Contents, err = render(unit.Name, unit.Contents)
		if err != nil {
			return nil, err
		}
	}

	for i := range ign.Systemd.Units {
		unit := &ign.Systemd.Units[i]
		unit.Contents, err = render(unit.Name, unit.Contents)
		if err != nil {
			return nil, err
		}
	}

	return ign, nil
}

func renderIgnition2_3(tmpl *sabakan.IgnitionTemplate, render renderFunc) (interface{}, error) {
	ign := new(ign23.Config)
	err := json.Unmarshal([]byte(tmpl.Template), ign)
	if err != nil {
		return nil, err
	}

	ign.Ignition.Version = "2.3.0"
	for i := range ign.Passwd.Groups {
		g := &ign.Passwd.Groups[i]
		pfx := fmt.Sprintf("passwd.groups[%d].", i)
		g.Name, err = render(pfx+"name", g.Name)
		if err != nil {
			return nil, err
		}
		g.PasswordHash, err = render(pfx+"passwordHash", g.PasswordHash)
		if err != nil {
			return nil, err
		}
	}
	for i := range ign.Passwd.Users {
		u := &ign.Passwd.Users[i]
		pfx := fmt.Sprintf("passwd.users[%d].", i)
		u.Name, err = render(pfx+"name", u.Name)
		if err != nil {
			return nil, err
		}
		if u.PasswordHash != nil {
			passwd, err := render(pfx+"passwordHash", *u.PasswordHash)
			if err != nil {
				return nil, err
			}
			u.PasswordHash = &passwd
		}
		for j, key := range u.SSHAuthorizedKeys {
			name := pfx + fmt.Sprintf("sshAuthorizedKeys[%d]", j)
			k, err := render(name, string(key))
			if err != nil {
				return nil, err
			}
			u.SSHAuthorizedKeys[j] = ign23.SSHAuthorizedKey(k)
		}
		u.Gecos, err = render(pfx+"gecos", u.Gecos)
		if err != nil {
			return nil, err
		}
		u.HomeDir, err = render(pfx+"homeDir", u.HomeDir)
		if err != nil {
			return nil, err
		}
		u.PrimaryGroup, err = render(pfx+"primaryGroup", u.PrimaryGroup)
		if err != nil {
			return nil, err
		}
		for j, group := range u.Groups {
			name := pfx + fmt.Sprintf("groups[%d]", j)
			g, err := render(name, string(group))
			if err != nil {
				return nil, err
			}
			u.Groups[j] = ign23.Group(g)
		}
		u.Shell, err = render(pfx+"shell", u.Shell)
		if err != nil {
			return nil, err
		}
	}

	for i := range ign.Storage.Files {
		file := &ign.Storage.Files[i]
		if file.Contents.Verification.Hash != nil {
			h, err := render(file.Path, *file.Contents.Verification.Hash)
			if err != nil {
				return nil, err
			}
			file.Contents.Verification.Hash = &h
		}
		if !strings.HasPrefix(file.Contents.Source, "data:") {
			file.Contents.Source, err = render(file.Path, file.Contents.Source)
			if err != nil {
				return nil, err
			}
			continue
		}

		// Render the file contents if embedded.
		d, err := dataurl.DecodeString(file.Contents.Source)
		if err != nil {
			return nil, err
		}
		rendered, err := render(file.Path, string(d.Data))
		if err != nil {
			return nil, err
		}
		file.Contents.Source = "data:," + dataurl.EscapeString(rendered)
	}

	for i := range ign.Networkd.Units {
		unit := &ign.Networkd.Units[i]
		unit.Contents, err = render(unit.Name, unit.Contents)
		if err != nil {
			return nil, err
		}
	}

	for i := range ign.Systemd.Units {
		unit := &ign.Systemd.Units[i]
		unit.Contents, err = render(unit.Name, unit.Contents)
		if err != nil {
			return nil, err
		}
	}

	return ign, nil
}
