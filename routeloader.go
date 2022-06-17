package main

import (
	"wv2/routes"

	"github.com/gorilla/mux"
)

func loadRoutes(r *mux.Router) {
	r.HandleFunc("/widgets/{id}", Route(routes.WidgetsCreateWidget))

	// Generate doctree api
	r.HandleFunc("/doctree", Route(routes.DocsGenerateDoctree))

	// Get document API
	r.HandleFunc(`/docs/{rest:[a-zA-Z0-9=\-\/]+}`, Route(routes.DocsGetDocument))

	// Admin schema fetch
	r.HandleFunc("/ap/schema", Route(routes.AdminGetSchema))

	// Get allowed tables
	r.HandleFunc("/ap/schema/allowed-tables", Route(routes.AdminGetAllowedTables))

	// Staff verification endpoint
	r.HandleFunc("/ap/newcat", Route(routes.AdminNewStaff))

	// QR code endpoint used by staff verify to show QR code to user
	r.HandleFunc("/qr/{hash}", Route(routes.AdminQRCode))

	// Staff login endpoint (for admin panel)
	r.HandleFunc("/ap/pouncecat", Route(routes.AdminStaffLogin))

	r.HandleFunc("/ap/shadowsight", Route(routes.AdminCheckSessionValid))

	r.HandleFunc("/ap/tables/{table_name}", Route(routes.AdminGetTable))
}
