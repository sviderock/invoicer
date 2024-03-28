package template

import (
	"database/sql"
	"fmt"
	"invoice-manager/main/internal/constants"
	pb "invoice-manager/main/proto"
	"log"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrIDNotFound = fmt.Errorf("ID not found")
)

type Template struct {
	data *pb.Template
}

func (t *Template) GetFileDirPath() string {
	path_split := strings.Split(t.data.Path, "/")
	return filepath.Join(path_split[:len(path_split)-1]...)
}

func (t *Template) GetOSPath() string {
	path_split := strings.Split(t.data.Path, "/")
	return filepath.Join(path_split...)
}

func (t *Template) UpdateName(name string, db *sql.DB) error {
	stmt, err := db.Prepare("UPDATE templates SET name = ? WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(name, t.data.Id)
	if err != nil {
		return err
	}

	t.data.Name = name
	return nil
}

type Templates struct {
	db *sql.DB

	insert_stmt, retrieve_stmt, list_stmt, delete_stmt *sql.Stmt
}

func (ts *Templates) Insert(template pb.Template) (*Template, error) {
	res, err := ts.insert_stmt.Exec(
		template.Name,
		template.Ext,
		template.Path,
		template.Size,
		template.Thumbnail,
		time.Now().Unix(),
		time.Now().Unix(),
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	new_template, err := ts.Retrieve(int(id))
	if err != nil {
		return nil, err
	}

	return &new_template, nil
}

func (ts *Templates) Retrieve(id int) (Template, error) {
	row := ts.retrieve_stmt.QueryRow(id)
	data := pb.Template{}
	err := row.Scan(
		&data.Id,
		&data.Name,
		&data.Ext,
		&data.Path,
		&data.Size,
		&data.Thumbnail,
		&data.CreatedAt,
		&data.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return Template{}, ErrIDNotFound
	}

	return Template{data: &data}, err
}

func (ts *Templates) List() ([]Template, error) {
	rows, err := ts.list_stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := []Template{}
	for rows.Next() {
		i := pb.Template{}
		err = rows.Scan(
			&i.Id,
			&i.Name,
			&i.Ext,
			&i.Path,
			&i.Size,
			&i.Thumbnail,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			log.Println(err)

			return nil, err
		}

		log.Println(i)
		data = append(data, Template{data: &i})
	}

	return data, nil
}

func (ts *Templates) Delete(id int) error {
	_, err := ts.delete_stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func NewTemplates() (*Templates, error) {
	db, err := sql.Open("sqlite3", constants.DB_FILE)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS templates (
			template_id INTEGER NOT NULL PRIMARY KEY,
			template_name VARCHAR NOT NULL,
			template_ext VARCHAR(10) NOT NULL,
			template_path TEXT NOT NULL,
			template_size INTEGER NOT NULL,
			template_thumbnail_path TEXT NOT NULL,
			template_created_at INTEGER NOT NULL,
			template_updated_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	insert_stmt, err := db.Prepare("INSERT INTO templates VALUES(NULL, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	retrieve_stmt, err := db.Prepare(`
		SELECT 
			template_id,
			template_name,
			template_ext,
			template_path,
			template_size,
			template_thumbnail_path,
			template_created_at,
			template_updated_at
		FROM templates
		WHERE template_id = ?;
	`)
	if err != nil {
		return nil, err
	}

	delete_stmt, err := db.Prepare("DELETE FROM templates WHERE template_id = ?")
	if err != nil {
		return nil, err
	}

	list_stmt, err := db.Prepare("SELECT * FROM templates ORDER BY template_id DESC")
	if err != nil {
		return nil, err
	}

	return &Templates{
		db:            db,
		insert_stmt:   insert_stmt,
		retrieve_stmt: retrieve_stmt,
		delete_stmt:   delete_stmt,
		list_stmt:     list_stmt,
	}, nil
}
