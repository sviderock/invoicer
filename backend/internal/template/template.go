package template

import (
	"database/sql"
	"fmt"
	"invoice-manager/main/internal/constants"
	"invoice-manager/main/internal/helpers"
	pb "invoice-manager/main/proto"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrIDNotFound = fmt.Errorf("ID not found")
)

type Template struct {
	data                  *pb.Template
	public_path           string
	public_thumbnail_path string
}

func (t *Template) GetPublicTemplatePath() string {
	return helpers.PublicUrlToFile(t.data.Path)
}

func (t *Template) GetPublicThumbnailPath() string {
	return helpers.PublicUrlToFile(t.data.Thumbnail)
}

type Templates struct {
	db *sql.DB

	insert_stmt, retrieve_stmt, list_stmt, delete_stmt, update_name_stmt *sql.Stmt
}

func (ts *Templates) Insert(template *pb.Template) (*Template, error) {
	res, err := ts.insert_stmt.Exec(
		template.Name,
		template.Ext,
		template.Size,
		template.Path,
		helpers.PublicUrlToFile(template.Path),
		template.Thumbnail,
		helpers.PublicUrlToFile(template.Thumbnail),
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
	template := Template{data: &pb.Template{}}
	err := row.Scan(
		&template.data.Id,
		&template.data.Name,
		&template.data.Ext,
		&template.data.Size,
		&template.data.Path,
		&template.public_path,
		&template.data.Thumbnail,
		&template.public_thumbnail_path,
		&template.data.CreatedAt,
		&template.data.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return template, ErrIDNotFound
	}

	return template, err
}

func (ts *Templates) List() ([]Template, error) {
	rows, err := ts.list_stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := []Template{}
	for rows.Next() {
		row := Template{data: &pb.Template{}}
		err = rows.Scan(
			&row.data.Id,
			&row.data.Name,
			&row.data.Ext,
			&row.data.Size,
			&row.data.Path,
			&row.public_path,
			&row.data.Thumbnail,
			&row.public_thumbnail_path,
			&row.data.CreatedAt,
			&row.data.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		data = append(data, row)
	}

	return data, nil
}

func (ts *Templates) Delete(id int) error {
	template, err := ts.Retrieve(id)
	if err != nil {
		return err
	}

	_, err = ts.delete_stmt.Exec(id)
	if err != nil {
		return err
	}

	err = os.Remove(template.data.Path)
	if err != nil {
		return err
	}

	err = os.Remove(template.data.Thumbnail)
	if err != nil {
		return err
	}

	return nil
}

func (ts *Templates) UpdateName(id int, new_name string) (*Template, error) {
	_, err := ts.update_name_stmt.Exec(new_name, id)
	if err != nil {
		return nil, err
	}

	updated_template, err := ts.Retrieve(id)
	if err != nil {
		return nil, err
	}

	return &updated_template, nil
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
			template_size INTEGER NOT NULL,
			template_private_path TEXT NOT NULL,
			template_public_path TEXT NOT NULL,
			template_private_thumbnail_path TEXT NOT NULL,
			template_public_thumbnail_path TEXT NOT NULL,
			template_created_at INTEGER NOT NULL,
			template_updated_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	insert_stmt, err := db.Prepare("INSERT INTO templates VALUES(NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	retrieve_stmt, err := db.Prepare(`
		SELECT 
			template_id,
			template_name,
			template_ext,
			template_size,
			template_private_path,
			template_public_path,
			template_private_thumbnail_path,
			template_public_thumbnail_path,
			template_created_at,
			template_updated_at
		FROM templates
		WHERE template_id = ?
		ORDER BY template_created_at ASC;
	`)
	if err != nil {
		return nil, err
	}

	delete_stmt, err := db.Prepare("DELETE FROM templates WHERE template_id = ?")
	if err != nil {
		return nil, err
	}

	list_stmt, err := db.Prepare("SELECT * FROM templates ORDER BY template_created_at ASC")
	if err != nil {
		return nil, err
	}

	update_name_stmt, err := db.Prepare(
		"UPDATE templates SET template_name = ? WHERE template_id = ?",
	)
	if err != nil {
		return nil, err
	}

	return &Templates{
		db:               db,
		insert_stmt:      insert_stmt,
		retrieve_stmt:    retrieve_stmt,
		delete_stmt:      delete_stmt,
		list_stmt:        list_stmt,
		update_name_stmt: update_name_stmt,
	}, nil
}
