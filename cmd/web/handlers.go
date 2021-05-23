package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/saepiae/contact/pkg/models"
)

type NewContact struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	MiddleName string `json:"middleName"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Address    string `json:"address"`
}

// Просто рандомная страница, которая возвращает некую html- страничку
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = ts.Execute(w, nil)
	if err != nil {
		app.serverError(w, err)
	}
}

// Возвращает контакт с указанным id
func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	c, err := app.contacts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	data, err := json.Marshal(c)
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "%v\n", string(data))
}

// Возвращает все контакты
func (app *application) allContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := app.contacts.FindAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(contacts)
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "%v\n", string(data))
}

// Добавляет новый контакт
func (app *application) createContact(w http.ResponseWriter, r *http.Request) {
	disabled, w := handlerAllowedMethod(w, r, http.MethodPost, app)
	if disabled {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var contact NewContact
	err := decoder.Decode(&contact)
	if err != nil {
		app.clientError(w, 400)
		return
	}

	id, err := app.contacts.Insert(contact.FirstName, contact.LastName, contact.MiddleName, contact.Phone, contact.Email, contact.Address)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/contact/get?id=%d", id), http.StatusSeeOther)
}

// Редактирует существующий контакт
func (app *application) editContact(w http.ResponseWriter, r *http.Request) {
	disabled, w := handlerAllowedMethod(w, r, http.MethodPut, app)
	if disabled {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var contact NewContact
	err = decoder.Decode(&contact)
	if err != nil {
		app.clientError(w, 400)
		return
	}

	id, err = app.contacts.Update(id, contact.FirstName, contact.LastName, contact.MiddleName, contact.Phone, contact.Email, contact.Address)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/contact/get?id=%d", id), http.StatusSeeOther)
}

// Удаляет контакт по указанному id
func (app *application) deleteContact(w http.ResponseWriter, r *http.Request) {
	disabled, w := handlerAllowedMethod(w, r, http.MethodDelete, app)
	if disabled {
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	_, err = app.contacts.Delete(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/contact/all"), http.StatusSeeOther)
}

func findDublicatedContacts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Список дублирующихся контактов"))
}

func handlerAllowedMethod(w http.ResponseWriter, r *http.Request, method string, app *application) (bool, http.ResponseWriter) {
	forbidden := r.Method != method
	w.Header().Set("Content-Type", "application/json")
	if forbidden {
		w.Header().Add("Allow", method)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
	return forbidden, w
}
