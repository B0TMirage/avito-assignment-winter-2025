package database

import "fmt"

func MigrateUP() {
	tx, err := DB.Begin()
	if err != nil {
		fmt.Println(err)
	}
	createUserTable := `CREATE TABLE IF NOT EXISTS users(
		id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		username VARCHAR(64) UNIQUE NOT NULL,
		password VARCHAR(70) NOT NULL,
		coins INT NOT NULL DEFAULT 0
	);`

	createMerchTable := `CREATE TABLE IF NOT EXISTS merch(
		id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		title VARCHAR(64) NOT NULL UNIQUE,
		price INT NOT NULL
	);`

	createTransactionTable := `CREATE TABLE IF NOT EXISTS transactions(
		id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		from_user_id INT NOT NULL,
		to_user_id INT NOT NULL,
		amount INT NOT NULL,

		CONSTRAINT from_fk FOREIGN KEY (from_user_id) REFERENCES users(id),
		CONSTRAINT to_fk FOREIGN KEY (to_user_id) REFERENCES users(id)
	);`

	createUserMerchTable := `CREATE TABLE IF NOT EXISTS users_merch(
		id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		user_id INT NOT NULL,
		merch_id INT NOT NULL,

		CONSTRAINT user_fk FOREIGN KEY (user_id) REFERENCES users(id),
		CONSTRAINT merch_fk FOREIGN KEY (merch_id) REFERENCES merch(id)
	);`

	tx.Exec(createUserTable)
	tx.Exec(createMerchTable)
	tx.Exec(createTransactionTable)
	tx.Exec(createUserMerchTable)

	tx.Exec("INSERT INTO merch(title, price) VALUES('t-shirt', 80) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('cup', 20) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('book', 50) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('pen', 10) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('powerbank', 200) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('hoody', 300) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('umbrella', 200) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('socks', 10) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('wallet', 50) ON CONFLICT (title) DO NOTHING;")
	tx.Exec("INSERT INTO merch(title, price) VALUES('pink-hoody', 500) ON CONFLICT (title) DO NOTHING;")

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
	}
}
