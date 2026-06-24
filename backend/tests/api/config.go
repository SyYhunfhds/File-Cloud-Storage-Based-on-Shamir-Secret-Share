package api

const (
	scheme = "http"
	server = "localhost:8000"
	base   = scheme + "://" + server + "/"

	registerURL = base + "v1/user"                  // POST
	loginURL    = base + "v1/auth/login"            // POST
	meURL       = base + "v1/protected/user/me"     // GET
	logoutURL   = base + "v1/protected/auth/logout" // GET

	submitURL     = base + "v1/protected/item/submit"       // POST multipart
	downloadURL   = base + "v1/protected/item/download"     // POST
	updateURL     = base + "v1/protected/item/update"       // POST
	viewApplyURL  = base + "v1/protected/item/view/require" // POST
	passAllURL    = base + "v1/protected/audit/pass/all"    // GET
	auditListURL  = base + "v1/protected/audit/list"        // GET
	refreshURL    = base + "v1/protected/share/refresh"     // POST
	getOneItemURL = base + "v1/protected/item"              // GET
)
