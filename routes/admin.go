package routes

import (
	"bytes"
	"fmt"
	"image/png"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wv2/types"
	"wv2/utils"

	"github.com/alexedwards/argon2id"
	"github.com/gorilla/mux"
	"github.com/happeens/xkcdpass"
	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp/totp"
)

func AdminGetSchema(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "GET" {
		w.Write([]byte(invalidMethod))
		return
	}

	schemaOtps := utils.SchemaFilter{}

	if r.URL.Query().Get("table_name") != "" {
		schemaOtps.TableName = r.URL.Query().Get("table_name")
	}

	res, err := utils.GetSchema(opts.Context, opts.DB, schemaOtps)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	bytes, err := json.Marshal(res)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func AdminGetAllowedTables(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "GET" {
		w.Write([]byte(invalidMethod))
		return
	}

	auth, err := utils.AuthorizeUser(utils.AuthRequest{
		UserID:  r.URL.Query().Get("user_id"),
		Token:   r.Header.Get("Authorization"),
		DevMode: opts.DevMode,
		Context: opts.Context,
		DB:      opts.DB,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	bytes, err := json.Marshal(auth.AllowedTables)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	w.Write(bytes)
}

// Staff verification endpoint
func AdminNewStaff(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "POST" {
		w.Write([]byte(invalidMethod))
		return
	}

	auth, err := utils.AuthorizeUser(utils.AuthRequest{
		UserID:  r.URL.Query().Get("user_id"),
		Token:   r.Header.Get("Authorization"),
		DevMode: opts.DevMode,
		Context: opts.Context,
		DB:      opts.DB,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if auth.Verified {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You are already verified"))
		return
	}

	if auth.Perms.Perm < 2 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You do not have permission to do this"))
		return
	}

	// Check code sent in request body
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	var verified bool

	if opts.DevMode {
		verified = utils.CheckCodeDev(r.URL.Query().Get("user_id"), string(body))
	} else {
		verified = utils.CheckCodeSecure(r.URL.Query().Get("user_id"), string(body))
	}

	if !verified {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid code"))
		return
	}

	newTotp, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Fates List Electrodragon Auth",
		AccountName: r.URL.Query().Get("user_id"),
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	newPass := xkcdpass.GenerateWithLength(6)

	newPassHashed, err := argon2id.CreateHash(newPass, argon2id.DefaultParams)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	_, err = opts.DB.Exec(opts.Context, "UPDATE users SET staff_verify_code = $1, staff_password = $2 WHERE user_id = $3", body, newPassHashed, r.URL.Query().Get("user_id"))

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	opts.Bot.GuildMemberRoleAdd(opts.MainServer, r.URL.Query().Get("user_id"), auth.Perms.ID)
	opts.Bot.GuildMemberRoleAdd(opts.StaffServer, r.URL.Query().Get("user_id"), auth.Perms.StaffID)

	imgHash := utils.RandString(512)

	imgQr, err := newTotp.Image(150, 150)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	// Convert imgQr to webp
	buf := new(bytes.Buffer)

	png.Encode(buf, imgQr)

	// Insert newTotp.Image() into redis
	opts.Redis.Set(opts.Context, imgHash, buf.Bytes(), 5*time.Minute)

	imageUrl := "https://corona.fateslist.xyz/qr/" + imgHash

	if opts.DevMode {
		imageUrl = "https://localhost:1800/qr/" + imgHash
	}

	data := types.NewStaff{
		Pass:      newPass,
		SharedKey: newTotp.Secret(),
		Image:     imageUrl,
	}

	json.NewEncoder(w).Encode(data)
}

// QR code endpoint used by staff verify to show QR code to user
func AdminQRCode(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Write([]byte(invalidMethod))
		return
	}

	hash := mux.Vars(r)["hash"]

	// Get image from redis
	img, err := opts.Redis.Get(opts.Context, hash).Result()

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	// Write image to response
	w.Header().Set("Content-Type", "image/png")
	w.Write([]byte(img))
}

// Staff login endpoint (for admin panel)
func AdminStaffLogin(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "POST" {
		w.Write([]byte(invalidMethod))
		return
	}

	auth, err := utils.AuthorizeUser(utils.AuthRequest{
		UserID:   r.URL.Query().Get("user_id"),
		Token:    r.Header.Get("Authorization"),
		TOTP:     r.Header.Get("Frostpaw-MFA"),
		Password: r.Header.Get("Frostpaw-Pass"),
		DevMode:  opts.DevMode,
		Context:  opts.Context,
		DB:       opts.DB,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if !auth.Verified {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You have not completed staff verification yet"))
		return
	}

	if !auth.PasswordLogin {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Password incorrect. Retry staff verification if you have not done it before"))
		return
	}

	if !auth.MFA {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("MFA incorrect. Retry staff verification if you have not done it before"))
		return
	}

	if auth.Perms.Perm < 2 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You do not have permission to do this"))
		return
	}

	session := utils.RandString(512)

	authSession := map[string]string{
		"user_id": r.URL.Query().Get("user_id"),
		"token":   r.Header.Get("Authorization"),
	}

	bytes, err := json.Marshal(authSession)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	opts.Redis.Set(opts.Context, session, bytes, time.Hour*2)

	w.Write([]byte(session))
}

func AdminCheckSessionValid(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	auth, err := utils.AuthorizeUser(utils.AuthRequest{
		UserID:    r.URL.Query().Get("user_id"),
		Token:     r.Header.Get("Authorization"),
		SessionID: r.Header.Get("Frostpaw-ID"),
		DevMode:   opts.DevMode,
		Context:   opts.Context,
		DB:        opts.DB,
		Redis:     opts.Redis,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if auth.Perms.Perm < 2 || !auth.SessionValidated {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You do not have permission to do this"))
		return
	}

	w.Write([]byte("OK"))
}

/* Accepts the following parameters

- user_id -> The user ID

- limit -> The number of results to return

- offset -> The number of results to skip

- search_by -> The field to search by

- search_val -> The value to search for

- count -> Whether to return the total number of results or the results themselves */
func AdminGetTable(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	auth, err := utils.AuthorizeUser(utils.AuthRequest{
		UserID:    r.URL.Query().Get("user_id"),
		Token:     r.Header.Get("Authorization"),
		SessionID: r.Header.Get("Frostpaw-ID"),
		DevMode:   opts.DevMode,
		Context:   opts.Context,
		DB:        opts.DB,
		Redis:     opts.Redis,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if auth.Perms.Perm < 2 || !auth.SessionValidated {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You do not have permission to do this"))
		return
	}

	tableName := mux.Vars(r)["table_name"]

	if len(auth.AllowedTables) > 0 {
		// We have a limitation on allowed tables, check if the table is allowed
		var allowedTable bool

		for _, table := range auth.AllowedTables {
			if table == tableName {
				allowedTable = true
				break
			}
		}

		if !allowedTable {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("You do not have permission to do this"))
			return
		}
	}

	// Get schema
	schema, err := utils.GetSchema(opts.Context, opts.DB, utils.SchemaFilter{
		TableName: tableName,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	if len(schema) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("This table does not exist"))
		return
	}

	countStr := r.URL.Query().Get("count")

	count := (countStr == "true" || countStr == "1")

	var limit int64 = -1
	var offset int64 = -1

	if !count {
		limit, _ = strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		offset, _ = strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
		limit = utils.Min(utils.Max(limit, 50), 50)
		offset = utils.Max(offset, 0)
	}

	fmt.Println(limit, offset)

	searchBy := r.URL.Query().Get("search_by")
	searchVal := r.URL.Query().Get("search_val")
	limitN := "2"  // Limit number
	offsetN := "3" // Offset number

	parseSql := func(sql string) (pgx.Rows, error) {
		if count {
			fmt.Println("Counting", sql)

			if offsetN == "2" {
				return opts.DB.Query(opts.Context, strings.Replace(sql, "SELECT *", "SELECT COUNT(*)", 1))
			}

			return opts.DB.Query(opts.Context, strings.Replace(sql, "SELECT *", "SELECT COUNT(*)", 1), searchVal)
		}

		if offsetN == "2" {
			return opts.DB.Query(opts.Context, sql+" LIMIT $"+limitN+" OFFSET $"+offsetN, limit, offset)
		}

		return opts.DB.Query(opts.Context, sql+" LIMIT $"+limitN+" OFFSET $"+offsetN, searchVal, limit, offset)
	}

	// Normal case (no search val or search by)

	var cols pgx.Rows

	if searchBy == "" || searchVal == "" {
		limitN, offsetN = "1", "2"
		cols, err = parseSql("SELECT * FROM " + tableName)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}
	} else {
		// Make sure the searchBy column is in the schema
		var found bool

		for _, col := range schema {
			if col.ColumnName == searchBy && !col.Secret {
				found = true
				break
			}
		}

		if !found {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("The search_by column does not exist"))
			return
		}

		// Handle all searchVal cases
		if searchVal == "null" {
			limitN, offsetN = "1", "2"
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text IS NULL")
		} else if strings.HasPrefix(searchVal, ">") {
			searchVal = strings.TrimPrefix(searchVal, ">")
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text > $1")
		} else if strings.HasPrefix(searchVal, "<") {
			searchVal = strings.TrimPrefix(searchVal, "<")
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text < $1")
		} else if strings.HasPrefix(searchVal, ">=") {
			searchVal = strings.TrimPrefix(searchVal, ">=")
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text >= $1")
		} else if strings.HasPrefix(searchVal, "<=") {
			searchVal = strings.TrimPrefix(searchVal, "<=")
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text <= $1")
		} else if strings.HasPrefix(searchVal, "!=") {
			searchVal = strings.TrimPrefix(searchVal, "!=")
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text != $1")
		} else if strings.HasPrefix(searchVal, "=") {
			searchVal = strings.TrimPrefix(searchVal, "=")
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text = $1")
		} else {
			searchVal = "%" + strings.TrimPrefix(searchVal, "@") + "%"
			cols, err = parseSql("SELECT * FROM " + tableName + " WHERE " + searchBy + "::text ILIKE $1")
		}

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}
	}

	if cols == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Something happened?"))
		return
	}

	defer cols.Close()

	if count {
		var count int64
		cols.Next()
		err = cols.Scan(&count)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatInt(count, 10)))
		return
	}

	var rows []map[string]any = []map[string]any{}

	var fieldDescrs = cols.FieldDescriptions()

	var colData []string = make([]string, len(fieldDescrs))

	for i, fieldDescr := range fieldDescrs {
		colData[i] = string(fieldDescr.Name)
	}

	fmt.Println(colData)

	for cols.Next() {
		var row map[string]any = make(map[string]any)

		vals, err := cols.Values()

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		for i, val := range vals {
			if valD, ok := val.([16]uint8); ok {
				val = fmt.Sprintf("%x-%x-%x-%x-%x", valD[0:4], valD[4:6], valD[6:8], valD[8:10], valD[10:16])
			}

			if valD, ok := val.(map[string]any); ok {
				valla, err := json.Marshal(valD)

				if err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(internalError))
					return
				}

				val = string(valla)
			}

			if valD, ok := val.(int64); ok {
				if valD > 9007199254740914 {
					val = strconv.FormatInt(valD, 10)
				}
			}

			row[colData[i]] = val
		}

		// Remove out secret columns
		for _, col := range schema {
			if col.Secret {
				delete(row, col.ColumnName)
			}
		}

		rows = append(rows, row)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rows)
}
