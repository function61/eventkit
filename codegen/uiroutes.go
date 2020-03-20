package codegen

import (
	"net/url"
	"strings"
)

type uiRouteSpec struct {
	Id   string `json:"id"`
	Path string `json:"path"`
}

// need to uppercase b/c tslint complains about pascal case
func (u *uiRouteSpec) TsOptsName() string {
	return strings.Title(u.Id) + "Opts"
}

func (u *uiRouteSpec) HasOpts() bool {
	return len(u.PathAndQueryOpts()) > 0
}

// "/accounts/{id}?token={token}" => ["id", "token"]
func (u *uiRouteSpec) PathAndQueryOpts() []string {
	return u.placeholders(u.Path)
}

// "/accounts/{id}?token={token}" => ["id"]
func (u *uiRouteSpec) PathPlaceholders() []string {
	return u.placeholders(u.PathWithoutQuery())
}

// "/accounts/{id}?token={token}" => ["token"]
func (u *uiRouteSpec) QueryPlaceholders() []string {
	return u.placeholders(u.Query())
}

func (u *uiRouteSpec) placeholders(pathOrQuery string) []string {
	placeholders := []string{}

	if matches := routePlaceholderParseRe.FindStringSubmatch(pathOrQuery); matches != nil {
		for _, match := range matches[1:] {
			placeholders = append(placeholders, removeBraces(match))
		}
	}

	return placeholders
}

func (u *uiRouteSpec) PathReJavaScript() string {
	// "/account/{account}/import_otp_token" => "/account/([^/]+)/import_otp_token"
	reString := routePlaceholderParseRe.ReplaceAllStringFunc(u.PathWithoutQuery(), func(_ string) string {
		return "([^/]+)"
	})

	// "/^\/account\/([^\/]+)\/import_otp_token$/"
	return "/^" + strings.ReplaceAll(reString, "/", `\/`) + "$/"
}

func (u *uiRouteSpec) PathWithoutQuery() string {
	// FIXME: trick to make it work with hashes
	if strings.HasPrefix(u.Path, "#") {
		parsedUrl, err := url.Parse(u.Path[1:])
		if err != nil {
			panic(err)
		}

		return "#" + parsedUrl.Path
	} else {
		parsedUrl, err := url.Parse(u.Path)
		if err != nil {
			panic(err)
		}

		return parsedUrl.Path
	}
}

func (u *uiRouteSpec) Query() string {
	// FIXME: trick to make it work with hashes
	parsedUrl, err := url.Parse(strings.TrimPrefix(u.Path, "#"))
	if err != nil {
		panic(err)
	}

	return parsedUrl.RawQuery
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
