package codegen

import (
	"strings"
)

type uiRouteSpec struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	QueryParams []struct {
		Key  string      `json:"key"`
		Type DatatypeDef `json:"type"`
	} `json:"query_params"`
}

// need to uppercase b/c tslint complains about pascal case
func (u *uiRouteSpec) TsOptsName() string {
	return strings.Title(u.Id) + "Opts"
}

func (u *uiRouteSpec) HasOpts() bool {
	return (len(u.PathPlaceholders()) + len(u.QueryParams)) > 0
}

// "/accounts/{id}?token={token}" => ["id"]
func (u *uiRouteSpec) PathPlaceholders() []string {
	keys := []string{}
	for _, match := range routePlaceholderParseRe.FindAllStringSubmatch(u.Path, -1) {
		keys = append(keys, match[1])
	}

	return keys
}

func (u *uiRouteSpec) PathReJavaScript() string {
	// "/account/{account}/import_otp_token" => "/account/([^/]+)/import_otp_token"
	reString := routePlaceholderParseRe.ReplaceAllStringFunc(u.Path, func(_ string) string {
		return "([^/]+)"
	})

	// "/^\/account\/([^\/]+)\/import_otp_token$/"
	return "/^" + strings.ReplaceAll(reString, "/", `\/`) + "$/"
}

func (u *uiRouteSpec) TsPath() string {
	return routePlaceholderParseRe.ReplaceAllStringFunc(u.Path, func(match string) string {
		// "{id}" => "id"
		placeholder := removeBraces(match)

		return "${encodeURIComponent(opts." + placeholder + ")}"
	})
}

func removeBraces(input string) string {
	return strings.Trim(input, "{}")
}
