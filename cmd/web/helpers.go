package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r), // Add the CSRF token.
	}
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request, in the same way that we did in our
	// snippetCreatePost handler.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call Decode() on our decoder instance, passing the target destination as
	// the first parameter.
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {

		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// For all other errors, we return them as normal.
		return err
	}

	return nil
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		// Use debug.Stack() to get the stack trace. This returns a byte slice, which
		// we need to convert to a string so that it's readable in the log entry.
		trace = string(debug.Stack())
	)

	// Include the trace in the log entry.
	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
