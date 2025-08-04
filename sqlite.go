package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Filename = ""
)

type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

func openConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", Filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// The function returns the User ID of the username and -1 if the user does not exist
func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println("Exists(), con:", err)
		return -1
	}
	defer db.Close()

	userID := -1
	stmt := fmt.Sprintf(`SELECT ID FROM Users where Username = '%s'`, username)
	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("exists() Scan", err)
			return -1
		}
		userID = id
	}
	return userID
}

// AddUser adds a new user to the database
// Returns new User ID
// -1 if err
func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()
	userID := exists(d.Username)
	if userID != -1 {
		fmt.Println("User already exists:", d.Username)
		return -1
	}
	insertStmt := `INSERT INTO Users values (NULL, ?)`
	_, err = db.Exec(insertStmt, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	userID = exists(d.Username)
	if userID == -1 {
		return userID
	}
	insertStmt = `INSERT INTO Userdata values (?,?,?,?)`
	_, err = db.Exec(insertStmt, userID, d.Name, d.Surname, d, d.Description)
	if err != nil {
		fmt.Println("db.Exec()", err)
		return -1
	}
	return userID
}

// DeleteUser deletes an existing user
func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	// Delete from Users
	res, err := db.Exec(`DELETE FROM Users WHERE ID = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("user %d not found", id)
	}

	// Delete from Userdata
	res, err = db.Exec(`DELETE FROM Userdata WHERE UserID = ?`, id)
	if err != nil {
		return err
	}
	n, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("user %d not found", id)
	}

	return nil
}

// ListUsers lists all users in the db
func ListUsers() ([]Userdata, error) {
	Data := []Userdata{}
	db, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT ID, Username, Name, Surname, Description FROM Users, Userdata WHERE Users.ID = Userdata.UserID`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var desc string
		err = rows.Scan(&id, &username, &name, &surname, &desc)
		temp := Userdata{ID: id, Name: username, Surname: surname, Description: desc}
		Data = append(Data, temp)
		if err != nil {
			return nil, err
		}
	}
	return Data, nil
}

// UpdateUser is for upd user
func UpdateUser(d Userdata) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID == -1 {
		return errors.New("User does not exist")
	}
	d.ID = userID
	updStmt := `UPDATE Userdata set Name = ?, Surname = ?, Description = ? where UserID = ?`
	res, err := db.Exec(updStmt, d.Name, d.Surname, d.Description, d.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("user %d not found or no changes", d.ID)
	}
	return nil
}
