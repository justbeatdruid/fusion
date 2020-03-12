package filter

import (
	//"strings"

	"github.com/chinamobile/nlpt/pkg/go-restful"
	"net/http"

	"github.com/chinamobile/nlpt/pkg/auth"
)

type TokenFilter struct {
}

func NewTokenFilter() *TokenFilter {
	return &TokenFilter{}
}

const tokenHeader = "X-auth-Token"

var tokenList = []string{
	"8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0",
	"0d52608f2ed610bdbc16e05c11a051a1cab86a19a3da94e9017ac68b62723c4b",
	"b291c86df58ea3ae8775f9a008ea08c26c2e9839b1a9051532a9b8688ab15835",
	"a16ce7b8c91fb85dc4b1b72a9f27e68f0a23a42b2aa7246869259f19623342bd",
	"b8672f91160c7f88157cd6ca457f09b07cb7fe4a7831b48e6b248200615d3311",
	"36c04a44fa28adaaaf7f1aeedd5842bd975cbf589d6dd9c7ce0f2d6f3ccb93da",
	"3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8",
	"71e2601dff5c544bd3d4976c997231ea5b1fe04c28cf4b14686dffd746cc742c",
	"4dd67c74d5b41e01ae178740c1da5d78da2fafeeba46356f9eceedf21cfca6f4",
	"ea2eacac787b57969a80525cf5b3028ca93d4a427878947421803dfc5162ed23",
	"fb0be75186d9cc23b1e6c7ba75f8e13708c98c05134464c2a57715a40bf3fe09",
	"97d04036e97fdd243810dca28b5fd39ecb57eedeab7e737233a8290d0aa17d20",
	"677f5bc56650cfbcac7e9ee94588bf1f698f369f4e5cdd5f11f3cd71b704583b",
	"35c2b85d63998bf6527e0cd5d4e55fde8199dfce18e94d91a9985aaebab887fa",
	"7d09038047ae330c27a16278ab609f3f2f00d2fb7a049725926a368cc8b835a3",
	"b0f05e76f2be4998c9fe4d7a2306833741b967f470ac2b8ed460a6b93b912850",
	"b5e0f9ce9bcac22753a6b07ba15b3d1f6e34b5d4c3a72cc00e9bb844ebde8976",
	"363c51354db36459734100899d68f31da0391cf7f10dd34f00d4b7d767fde849",
	"941a56da52c163ba18517124673aedac7b897388f80ff1909029cb9bec6558db",
	"e622ade7696a096a010ee1537257b3be6f9d197f1804dfd9582dfcd9a84ea704",
	"33cb06d62d16dee8646daa3fd18ed20f938f1aca4b96e1555907a6aa14550faf",
}

// ported from go-restful/options_filter.go
// add allowed all headers, credentials
func (o *TokenFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token := req.HeaderParameter("X-auth-Token")
	for _, t := range tokenList {
		if t == token {
			auth.SetAuthUser(req, auth.AuthUser{
				Name:      o.getUserName(req),
				Namespace: o.getUserNamespace(req),
			})
			chain.ProcessFilter(req, resp)
			return
		}
	}
	resp.WriteHeader(http.StatusUnauthorized)
}

func (o *TokenFilter) getUserName(req *restful.Request) string {
	uid := req.HeaderParameter("userId")
	if len(uid) > 0 {
		return uid
	}
	u := req.HeaderParameter("user")
	if len(u) > 0 {
		return u
	}
	return "admin"
}

func (o *TokenFilter) getUserNamespace(req *restful.Request) string {
	u := req.HeaderParameter("tenantId")
	if len(u) > 0 {
		return u
	}
	return "default"
}
