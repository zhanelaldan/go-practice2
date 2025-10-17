package main

import (
    "fmt"
    "log"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

// структура таблицы users 
type User struct {
    ID      int     `db:"id"`
    Name    string  `db:"name"`
    Email   string  `db:"email"`
    Balance float64 `db:"balance"`
}

// подключение к базе данных 
func main() {
    dsn := "postgres://user:password@localhost:5430/mydatabase?sslmode=disable"

    db, err := sqlx.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Connection error:", err)
    }
    defer db.Close()

    // настройки пула соединений
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        log.Fatal("Ping error:", err)
    }
    fmt.Println("✅ Connected to PostgreSQL successfully!")

    // добавляем двух пользователей для примера
    u1 := User{Name: "Alice", Email: "alice@mail.com", Balance: 100.0}
    u2 := User{Name: "Bob", Email: "bob@mail.com", Balance: 200.0}
    InsertUser(db, u1)
    InsertUser(db, u2)

    // выводим всех пользователей 
    fmt.Println("\n📋 All users before transfer:")
    users, _ := GetAllUsers(db)
    for _, user := range users {
        fmt.Printf("ID: %d | %s | %.2f\n", user.ID, user.Name, user.Balance)
    }

    // получаем одного пользователя 
    fmt.Println("\n🔍 Get user by ID (1):")
    user, err := GetUserByID(db, 1)
    if err != nil {
        log.Println("GetUserByID error:", err)
    } else {
        fmt.Printf("Found: ID=%d | Name=%s | Balance=%.2f\n", user.ID, user.Name, user.Balance)
    }

    // транзакция перевода 
    fmt.Println("\n💸 Transferring 50.0 from Alice (1) → Bob (2)...")
    err = TransferBalance(db, 1, 2, 50.0)
    if err != nil {
        log.Println("Transfer error:", err)
    } else {
        fmt.Println("✅ Transfer successful!")
    }

    // проверяем результат 
    fmt.Println("\n📊 All users after transfer:")
    users, _ = GetAllUsers(db)
    for _, user := range users {
        fmt.Printf("ID: %d | %s | %.2f\n", user.ID, user.Name, user.Balance)
    }
}

// вставка нового пользователя 
func InsertUser(db *sqlx.DB, user User) error {
    query := `INSERT INTO users (name, email, balance) VALUES (:name, :email, :balance)`
    _, err := db.NamedExec(query, user)
    return err
}

// получить всех пользователей 
func GetAllUsers(db *sqlx.DB) ([]User, error) {
    var users []User
    err := db.Select(&users, "SELECT * FROM users ORDER BY id")
    return users, err
}

// получить одного пользователя по id
func GetUserByID(db *sqlx.DB, id int) (User, error) {
    var user User
    err := db.Get(&user, "SELECT * FROM users WHERE id=$1", id)
    return user, err
}

// перевод денег 
func TransferBalance(db *sqlx.DB, fromID int, toID int, amount float64) error {
    tx, err := db.Beginx()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }

    // проверка отправителя
    var sender User
    err = tx.Get(&sender, "SELECT * FROM users WHERE id=$1", fromID)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("sender not found: %v", err)
    }

    if sender.Balance < amount {
        tx.Rollback()
        return fmt.Errorf("insufficient funds for user ID %d", fromID)
    }

    // списание
    _, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id=$2", amount, fromID)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to withdraw: %v", err)
    }

    // зачисление
    _, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id=$2", amount, toID)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to deposit: %v", err)
    }

    // подтверждение транзакции
    err = tx.Commit()
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to commit: %v", err)
    }

    return nil
}
