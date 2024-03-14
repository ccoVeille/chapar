package restful

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/mirzakhany/chapar/internal/domain"
	"github.com/mirzakhany/chapar/ui/pages/requests/component"
	"github.com/mirzakhany/chapar/ui/widgets"
)

type Auth struct {
	DropDown *widgets.DropDown

	auth *domain.Auth

	TokenForm *component.Form
	BasicForm *component.Form

	onChange func(auth *domain.Auth)
}

func NewAuth(auth *domain.Auth) *Auth {
	a := &Auth{
		auth: auth,
		DropDown: widgets.NewDropDown(
			widgets.NewDropDownOption("None"),
			widgets.NewDropDownOption("Basic"),
			widgets.NewDropDownOption("Token"),
		),

		TokenForm: component.NewForm([]*component.Field{
			{Label: "Token", Value: ""},
		}),
		BasicForm: component.NewForm([]*component.Field{
			{Label: "Username", Value: ""},
			{Label: "Password", Value: ""},
		}),
	}

	if auth == nil {
		auth = &domain.Auth{}
	}

	a.DropDown.SetSelectedByValue(auth.Type)
	if auth.BasicAuth != nil {
		a.BasicForm.SetValues(map[string]string{
			"Username": auth.BasicAuth.Username,
			"Password": auth.BasicAuth.Password,
		})
	}

	if auth.TokenAuth != nil {
		a.TokenForm.SetValues(map[string]string{
			"Token": auth.TokenAuth.Token,
		})
	}

	return a
}

func (a *Auth) SetOnChange(f func(auth *domain.Auth)) {
	a.onChange = f

	a.DropDown.SetOnValueChanged(func(selected string) {
		a.auth.Type = selected
		a.onChange(a.auth)
	})

	a.TokenForm.SetOnChange(func(values map[string]string) {
		a.auth.TokenAuth.Token = values["Token"]
		a.onChange(a.auth)
	})

	a.BasicForm.SetOnChange(func(values map[string]string) {
		a.auth.BasicAuth.Username = values["Username"]
		a.auth.BasicAuth.Password = values["Password"]
		a.onChange(a.auth)
	})
}

func (a *Auth) Layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
	}.Layout(gtx,
		layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.DropDown.Layout(gtx, theme)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			switch a.DropDown.GetSelected().Text {
			case "Token":
				return a.TokenForm.Layout(gtx, theme)
			case "Basic":
				return a.BasicForm.Layout(gtx, theme)
			default:
				return layout.Dimensions{}
			}
		}),
	)
}
