package plugins

import (
	"errors"
	"strings"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
)

func App(pluginKey string, owner string, cv *assets.App) *database.App {
	c := &database.App{
		Details: database.Details{
			Name:        &cv.Name,
			Description: cv.Description,
			Icon:        cv.Icon,
		},
		Enabled: cv.Enabled,
		Plugin:  &pluginKey,
		Owner:   &owner,
	}
	if cv.Scopes != nil {
		c.Scopes = &database.AppScopeArray{
			ScopeArray: database.ScopeArray{
				Scopes: *cv.Scopes,
			},
		}
	}
	if cv.AccessToken == nil || !(*cv.AccessToken) {
		empty := ""
		c.AccessToken = &empty
	}
	if cv.SettingsSchema != nil {
		jo := database.JSONObject(*cv.SettingsSchema)
		c.SettingsSchema = &jo
	}
	if cv.Settings != nil {
		jo := database.JSONObject(*cv.Settings)
		c.Settings = &jo
	}
	if cv.Type != nil {
		c.Type = cv.Type
	}
	return c
}

func AppObject(app string, key string, as *assets.Object) *database.Object {
	s := &database.Object{
		Details: database.Details{
			Name:        &as.Name,
			Description: as.Description,
			Icon:        as.Icon,
		},
		App:  &app,
		Key:  &key,
		Type: &as.Type,
	}
	if as.Meta != nil {
		jo := database.JSONObject(*as.Meta)
		s.Meta = &jo
	}
	if as.Scopes != nil {
		s.Scopes = &database.ScopeArray{
			Scopes: *as.Scopes,
		}
	}

	return s
}

func CreateApp(c *rest.Context, owner string, pluginKey string) (string, string, error) {
	if c.DB.Type() != database.UserType && c.DB.Type() != database.AdminType {
		return "", "", database.ErrAccessDenied("Only users can create apps")
	}
	if c.DB.Type() == database.UserType && owner == "" {
		owner = c.DB.ID()
	}
	if owner == "" {
		return "", "", errors.New("App must have an owner")
	}
	pk := strings.Split(pluginKey, ":")
	if len(pk) != 2 {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}
	adb := c.DB.AdminDB()
	a := adb.Assets()

	p, ok := a.Config.Plugins[pk[0]]
	if !ok {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}

	app, ok := p.Apps[pk[1]]
	if !ok {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}

	// Check if this key is from an *active* plugin

	ap := a.Config.GetActivePlugins()
	hadPlugin := false
	for _, p := range ap {
		if p == pk[0] {
			hadPlugin = true
			break
		}
	}
	if !hadPlugin {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}

	if app.Unique != nil && *app.Unique {
		noicon := false
		a, err := adb.ListApps(&database.ListAppOptions{
			Plugin: &pluginKey,
			User:   &owner,
			Icon:   &noicon,
		})
		if err != nil {
			return "", "", err
		}
		if len(a) >= 1 {
			return "", "", errors.New("This unique plugin app already exists")
		}
	}

	aid, akey, err := adb.CreateApp(App(pluginKey, owner, app))
	if err != nil {
		return aid, akey, err
	}
	for skey, sv := range app.Objects {
		// We perform the next stuff as admin
		if sv.AutoCreate == nil || *sv.AutoCreate == true {
			_, err := c.Request(c, "POST", "/api/heedy/v1/objects", AppObject(aid, skey, sv), map[string]string{"X-Heedy-As": "heedy"})
			if err != nil {
				adb.DelApp(aid)
				return "", "", err
			}
		}

	}

	return aid, akey, err
}